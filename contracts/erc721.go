package token

import (
    "encoding/json"
    "fmt"
    "github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

const balancePrefix = "balance"
const nftPrefix = "nft"
const approvalPrefix = "approval"
const nameKey1 = "name"
const symbolKey1 = "symbol"

type Nft struct {
    TokenId  string `json:"tokenId"`
    Owner    string `json:"owner"`
    TokenURI string `json:"tokenURI"`
    Approved string `json:"approved"`
}

type Approval struct {
    Owner    string `json:"owner"`
    Operator string `json:"operator"`
    Approved bool   `json:"approved"`
}

type Transfer struct {
    From    string `json:"from"`
    To      string `json:"to"`
    TokenId string `json:"tokenId"`
}

type TokenERC721Contract struct {
    kalpsdk.Contract
}

func _readNFT(ctx kalpsdk.TransactionContextInterface, tokenId string) (*Nft, error) {
    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        return nil, fmt.Errorf("failed to CreateCompositeKey %s: %v", tokenId, err)
    }

    nftBytes, err := ctx.GetState(nftKey)
    if err != nil {
        return nil, fmt.Errorf("failed to GetState %s: %v", tokenId, err)
    }

    nft := new(Nft)
    err = json.Unmarshal(nftBytes, nft)
    if err != nil {
        return nil, fmt.Errorf("failed to Unmarshal nftBytes: %v", err)
    }

    return nft, nil
}

func _nftExists(ctx kalpsdk.TransactionContextInterface, tokenId string) bool {
    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        panic("error creating CreateCompositeKey:" + err.Error())
    }

    nftBytes, err := ctx.GetState(nftKey)
    if err != nil {
        panic("error GetState nftBytes:" + err.Error())
    }

    return len(nftBytes) > 0
}

func (c *TokenERC721Contract) BalanceOf(ctx kalpsdk.TransactionContextInterface, owner string) int {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        panic("failed to check if contract is already initialized:" + err.Error())
    }
    if !initialized {
        panic("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    iterator, err := ctx.GetStateByPartialCompositeKey(balancePrefix, []string{owner})
    if err != nil {
        panic("Error creating asset chaincode:" + err.Error())
    }

    balance := 0
    for iterator.HasNext() {
        _, err := iterator.Next()
        if err != nil {
            return 0
        }
        balance++
    }
    return balance
}
func (c *TokenERC721Contract) OwnerOf(ctx kalpsdk.TransactionContextInterface, tokenId string) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return "", fmt.Errorf("could not process OwnerOf for tokenId: %w", err)
    }

    return nft.Owner, nil
}

func (c *TokenERC721Contract) Approve(ctx kalpsdk.TransactionContextInterface, operator string, tokenId string) (bool, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return false, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    sender, err := ctx.GetUserID()
    if err != nil {
        return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return false, fmt.Errorf("failed to _readNFT: %v", err)
    }

    owner := nft.Owner
    operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
    if err != nil {
        return false, fmt.Errorf("failed to get IsApprovedForAll: %v", err)
    }
    if owner != sender && !operatorApproval {
        return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
    }

    nft.Approved = operator
    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey %s: %v", nftKey, err)
    }

    nftBytes, err := json.Marshal(nft)
    if err != nil {
        return false, fmt.Errorf("failed to marshal nftBytes: %v", err)
    }

    err = ctx.PutStateWithoutKYC(nftKey, nftBytes)
    if err != nil {
        return false, fmt.Errorf("failed to PutState for nftKey: %v", err)
    }

    return true, nil
}

func (c *TokenERC721Contract) SetApprovalForAll(ctx kalpsdk.TransactionContextInterface, operator string, approved bool) (bool, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return false, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    sender, err := ctx.GetUserID()
    if err != nil {
        return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
    }

    nftApproval := new(Approval)
    nftApproval.Owner = sender
    nftApproval.Operator = operator
    nftApproval.Approved = approved

    approvalKey, err := ctx.CreateCompositeKey(approvalPrefix, []string{sender, operator})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
    }

    approvalBytes, err := json.Marshal(nftApproval)
    if err != nil {
        return false, fmt.Errorf("failed to marshal approvalBytes: %v", err)
    }

    err = ctx.PutStateWithoutKYC(approvalKey, approvalBytes)
    if err != nil {
        return false, fmt.Errorf("failed to PutState approvalBytes: %v", err)
    }

    err = ctx.SetEvent("ApprovalForAll", approvalBytes)
    if err != nil {
        return false, fmt.Errorf("failed to SetEvent ApprovalForAll: %v", err)
    }

    return true, nil
}


func (c *TokenERC721Contract) IsApprovedForAll(ctx kalpsdk.TransactionContextInterface, owner string, operator string) (bool, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return false, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    approvalKey, err := ctx.CreateCompositeKey(approvalPrefix, []string{owner, operator})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
    }
    approvalBytes, err := ctx.GetState(approvalKey)
    if err != nil {
        return false, fmt.Errorf("failed to GetState approvalBytes %s: %v", approvalBytes, err)
    }

    if len(approvalBytes) < 1 {
        return false, nil
    }

    approval := new(Approval)
    err = json.Unmarshal(approvalBytes, approval)
    if err != nil {
        return false, fmt.Errorf("failed to Unmarshal: %v, string %s", err, string(approvalBytes))
    }

    return approval.Approved, nil
}

func (c *TokenERC721Contract) GetApproved(ctx kalpsdk.TransactionContextInterface, tokenId string) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "false", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "false", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return "false", fmt.Errorf("failed GetApproved for tokenId : %v", err)
    }
    return nft.Approved, nil
}

func (c *TokenERC721Contract) TransferFrom(ctx kalpsdk.TransactionContextInterface, from string, to string, tokenId string) (bool, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return false, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    sender, err := ctx.GetUserID()
    if err != nil {
        return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return false, fmt.Errorf("failed to _readNFT : %v", err)
    }

    owner := nft.Owner
    operator := nft.Approved
    operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
    if err != nil {
        return false, fmt.Errorf("failed to get IsApprovedForAll : %v", err)
    }
    if owner != sender && operator != sender && !operatorApproval {
        return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
    }

    if owner != from {
        return false, fmt.Errorf("the from is not the current owner")
    }

    nft.Approved = ""
    nft.Owner = to
    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
    }

    nftBytes, err := json.Marshal(nft)
    if err != nil {
        return false, fmt.Errorf("failed to marshal approval: %v", err)
    }

    err = ctx.PutStateWithoutKYC(nftKey, nftBytes)
    if err != nil {
        return false, fmt.Errorf("failed to PutState nftBytes %s: %v", nftBytes, err)
    }

    balanceKeyFrom, err := ctx.CreateCompositeKey(balancePrefix, []string{from, tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey from: %v", err)
    }

    err = ctx.DelStateWithoutKYC(balanceKeyFrom)
    if err != nil {
        return false, fmt.Errorf("failed to DelState balanceKeyFrom %s: %v", nftBytes, err)
    }

    balanceKeyTo, err := ctx.CreateCompositeKey(balancePrefix, []string{to, tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey to: %v", err)
    }
    err = ctx.PutStateWithoutKYC(balanceKeyTo, []byte{0})
    if err != nil {
        return false, fmt.Errorf("failed to PutState balanceKeyTo %s: %v", balanceKeyTo, err)
    }

    transferEvent := new(Transfer)
    transferEvent.From = from
    transferEvent.To = to
    transferEvent.TokenId = tokenId

    transferEventBytes, err := json.Marshal(transferEvent)
    if err != nil {
        return false, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
    }

    err = ctx.SetEvent("Transfer", transferEventBytes)
    if err != nil {
        return false, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
    }
    return true, nil
}

func (c *TokenERC721Contract) Name(ctx kalpsdk.TransactionContextInterface) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    bytes, err := ctx.GetState(nameKey1)
    if err != nil {
        return "", fmt.Errorf("failed to get Name bytes: %s", err)
    }

    return string(bytes), nil
}

func (c *TokenERC721Contract) Symbol(ctx kalpsdk.TransactionContextInterface) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    bytes, err := ctx.GetState(symbolKey1)
    if err != nil {
        return "", fmt.Errorf("failed to get Symbol: %v", err)
    }

    return string(bytes), nil
}

func (c *TokenERC721Contract) TokenURI(ctx kalpsdk.TransactionContextInterface, tokenId string) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return "", fmt.Errorf("failed to get TokenURI: %v", err)
    }
    return nft.TokenURI, nil
}

func (c *TokenERC721Contract) TotalSupply(ctx kalpsdk.TransactionContextInterface) int {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        panic("failed to check if contract is already initialized:" + err.Error())
    }
    if !initialized {
        panic("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    iterator, err := ctx.GetStateByPartialCompositeKey(nftPrefix, []string{})
    if err != nil {
        panic("Error creating GetStateByPartialCompositeKey:" + err.Error())
    }

    totalSupply := 0
    for iterator.HasNext() {
        _, err := iterator.Next()
        if err != nil {
            return 0
        }
        totalSupply++
    }
    return totalSupply
}

func (c *TokenERC721Contract) Initialize(ctx kalpsdk.TransactionContextInterface, name string, symbol string) (bool, error) {
    clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return false, fmt.Errorf("failed to get clientMSPID: %v", err)
    }
    if clientMSPID != "mailabs" {
        return false, fmt.Errorf("client is not authorized to set the name and symbol of the token")
    }

    bytes, err := ctx.GetState(nameKey1)
    if err != nil {
        return false, fmt.Errorf("failed to get Name: %v", err)
    }
    if bytes != nil {
        return false, fmt.Errorf("contract options are already set, client is not authorized to change them")
    }

    err = ctx.PutStateWithoutKYC(nameKey1, []byte(name))
    if err != nil {
        return false, fmt.Errorf("failed to PutState nameKey1 %s: %v", nameKey1, err)
    }

    err = ctx.PutStateWithoutKYC(symbolKey1, []byte(symbol))
    if err != nil {
        return false, fmt.Errorf("failed to PutState symbolKey1 %s: %v", symbolKey1, err)
    }

    return true, nil
}
func (c *TokenERC721Contract) MintWithTokenURI(ctx kalpsdk.TransactionContextInterface, tokenId string, tokenURI string) (*Nft, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return nil, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return nil, fmt.Errorf("failed to get clientMSPID: %v", err)
    }

    if clientMSPID != "mailabs" {
        return nil, fmt.Errorf("client is not authorized to set the name and symbol of the token")
    }

    minter, err := ctx.GetUserID()
    if err != nil {
        return nil, fmt.Errorf("failed to get minter id: %v", err)
    }

    exists := _nftExists(ctx, tokenId)
    if exists {
        return nil, fmt.Errorf("the token %s is already minted.: %v", tokenId, err)
    }

    nft := new(Nft)
    nft.TokenId = tokenId
    nft.Owner = minter
    nft.TokenURI = tokenURI

    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        return nil, fmt.Errorf("failed to CreateCompositeKey to nftKey: %v", err)
    }

    nftBytes, err := json.Marshal(nft)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal nft: %v", err)
    }

    err = ctx.PutStateWithoutKYC(nftKey, nftBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to PutState nftBytes %s: %v", nftBytes, err)
    }

    balanceKey, err := ctx.CreateCompositeKey(balancePrefix, []string{minter, tokenId})
    if err != nil {
        return nil, fmt.Errorf("failed to CreateCompositeKey to balanceKey: %v", err)
    }

    err = ctx.PutStateWithoutKYC(balanceKey, []byte{'\u0000'})
    if err != nil {
        return nil, fmt.Errorf("failed to PutState balanceKey %s: %v", nftBytes, err)
    }

    transferEvent := new(Transfer)
    transferEvent.From = "0x0"
    transferEvent.To = minter
    transferEvent.TokenId = tokenId

    transferEventBytes, err := json.Marshal(transferEvent)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
    }

    err = ctx.SetEvent("Transfer", transferEventBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
    }

    return nft, nil
}
func (c *TokenERC721Contract) Burn(ctx kalpsdk.TransactionContextInterface, tokenId string) (bool, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return false, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    owner, err := ctx.GetUserID()
    if err != nil {
        return false, fmt.Errorf("failed to GetClientIdentity owner64: %v", err)
    }

    nft, err := _readNFT(ctx, tokenId)
    if err != nil {
        return false, fmt.Errorf("failed to _readNFT nft : %v", err)
    }
    if nft.Owner != owner {
        return false, fmt.Errorf("non-fungible token %s is not owned by %s", tokenId, owner)
    }

    nftKey, err := ctx.CreateCompositeKey(nftPrefix, []string{tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey tokenId: %v", err)
    }

    err = ctx.DelStateWithoutKYC(nftKey)
    if err != nil {
        return false, fmt.Errorf("failed to DelState nftKey: %v", err)
    }

    balanceKey, err := ctx.CreateCompositeKey(balancePrefix, []string{owner, tokenId})
    if err != nil {
        return false, fmt.Errorf("failed to CreateCompositeKey balanceKey %s: %v", balanceKey, err)
    }

    err = ctx.DelStateWithoutKYC(balanceKey)
    if err != nil {
        return false, fmt.Errorf("failed to DelState balanceKey %s: %v", balanceKey, err)
    }

    transferEvent := new(Transfer)
    transferEvent.From = owner
    transferEvent.To = "0x0"
    transferEvent.TokenId = tokenId

    transferEventBytes, err := json.Marshal(transferEvent)
    if err != nil {
        return false, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
    }

    err = ctx.SetEvent("Transfer", transferEventBytes)
    if err != nil {
        return false, fmt.Errorf("failed to SetEvent transferEventBytes: %v", err)
    }

    return true, nil
}
func (c *TokenERC721Contract) ClientAccountBalance(ctx kalpsdk.TransactionContextInterface) (int, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return 0, fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return 0, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    clientAccountID, err := ctx.GetUserID()
    if err != nil {
        return 0, fmt.Errorf("failed to GetClientIdentity minter: %v", err)
    }

    return c.BalanceOf(ctx, clientAccountID), nil
}

func (c *TokenERC721Contract) ClientAccountID(ctx kalpsdk.TransactionContextInterface) (string, error) {
    initialized, err := checkInitialized1(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
    }
    if !initialized {
        return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
    }

    clientAccount, err := ctx.GetUserID()
    if err != nil {
        return "", fmt.Errorf("failed to GetClientIdentity minter: %v", err)
    }

    return clientAccount, nil
}


func checkInitialized1(ctx kalpsdk.TransactionContextInterface) (bool, error) {
    tokenName, err := ctx.GetState(nameKey1)
    if err != nil {
        return false, fmt.Errorf("failed to get token name: %v", err)
    }
    if tokenName == nil {
        return false, nil
    }
    return true, nil
}