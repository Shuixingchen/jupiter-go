package solana

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

const defaultMaxRetries = uint(20)

type TxID string

type TokenAccount struct {
	Amount   decimal.Decimal
	Decimals uint8
}

type client struct {
	maxRetries uint
	clientRPC  rpcService
	wallet     Wallet
}

func newClient(
	wallet Wallet,
	rpcEndpoint string,
	opts ...ClientOption,
) (*client, error) {
	c := &client{
		maxRetries: defaultMaxRetries,
		wallet:     wallet,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("could not apply option: %w", err)
		}
	}

	if c.clientRPC == nil {
		if rpcEndpoint == "" {
			return nil, fmt.Errorf("rpcEndpoint is required when no RPC service is provided")
		}

		rpcClient := rpc.New(rpcEndpoint)
		c.clientRPC = rpcClient
	}

	return c, nil
}

// NewClient creates a new Solana client with the given wallet and RPC endpoint.
// If you want to monitor your transactions using a websocket endpoint, use NewClientWithWS.
func NewClient(
	wallet Wallet,
	rpcEndpoint string,
	opts ...ClientOption,
) (Client, error) {
	return newClient(wallet, rpcEndpoint, opts...)
}

// SendTransactionOnChain sends a transaction on-chain.
func (e client) SendTransactionOnChain(ctx context.Context, txBase64 string) (TxID, error) {
	latestBlockhash, err := e.clientRPC.GetLatestBlockhash(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get latest blockhash: %w", err)
	}

	tx, err := NewTransactionFromBase64(txBase64)
	if err != nil {
		return "", fmt.Errorf("could not deserialize swap transaction: %w", err)
	}

	tx.Message.RecentBlockhash = latestBlockhash.Value.Blockhash

	tx, err = e.wallet.SignTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("could not sign swap transaction: %w", err)
	}

	sig, err := e.clientRPC.SendTransactionWithOpts(ctx, &tx, rpc.TransactionOpts{
		MaxRetries:          &e.maxRetries,
		MinContextSlot:      &latestBlockhash.Context.Slot,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return "", fmt.Errorf("could not send transaction: %w", err)
	}

	return TxID(sig.String()), nil
}

// CheckSignature checks if a transaction with the given signature has been confirmed on-chain.
func (e client) CheckSignature(ctx context.Context, tx TxID) (bool, error) {
	sig, err := solana.SignatureFromBase58(string(tx))
	if err != nil {
		return false, fmt.Errorf("could not convert signature from base58: %w", err)
	}

	status, err := e.clientRPC.GetSignatureStatuses(ctx, false, sig)
	if err != nil {
		return false, fmt.Errorf("could not get signature status: %w", err)
	}

	if len(status.Value) == 0 {
		return false, fmt.Errorf("could not confirm transaction: no valid status")
	}

	if status.Value[0] == nil || status.Value[0].ConfirmationStatus != rpc.ConfirmationStatusFinalized {
		return false, fmt.Errorf("transaction not finalized yet")
	}

	if status.Value[0].Err != nil {
		return true, fmt.Errorf("transaction confirmed with error: %s", status.Value[0].Err)
	}

	return true, nil
}

// GetTokenAccountBalance returns the balance of an SPL token account.
func (e client) GetTokenAccountBalance(ctx context.Context, tokenAccount string) (TokenAccount, error) {
	tokenAccountPk, err := solana.PublicKeyFromBase58(tokenAccount)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("could not parse token account public key: %w", err)
	}

	resp, err := e.clientRPC.GetTokenAccountBalance(ctx, tokenAccountPk, rpc.CommitmentFinalized)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("could not get token account balance: %w", err)
	}

	if resp.Value == nil {
		return TokenAccount{}, fmt.Errorf("could not get token account balance: response value is nil")
	}

	value, err := decimal.NewFromString(resp.Value.Amount)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("could not convert token account balance to decimal: %w", err)
	}

	return TokenAccount{
		Amount:   value,
		Decimals: resp.Value.Decimals,
	}, nil
}

// GetSPLTokenBalance 获取指定钱包地址下特定SPL代币的余额
func (e client) GetSPLTokenBalance(ctx context.Context, walletAddress, tokenMint string) (TokenAccount, error) {
	// 将钱包地址转换为PublicKey
	walletPk, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("无法解析钱包地址: %w", err)
	}

	// 将代币Mint地址转换为PublicKey
	mintPk, err := solana.PublicKeyFromBase58(tokenMint)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("无法解析代币Mint地址: %w", err)
	}

	// 获取关联代币账户地址
	tokenAccountPk, _, err := solana.FindAssociatedTokenAddress(walletPk, mintPk)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("无法获取关联代币账户地址: %w", err)
	}

	// 获取代币账户余额
	resp, err := e.clientRPC.GetTokenAccountBalance(ctx, tokenAccountPk, rpc.CommitmentFinalized)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("无法获取代币账户余额: %w", err)
	}

	if resp.Value == nil {
		return TokenAccount{}, fmt.Errorf("获取代币账户余额失败: 返回值为空")
	}

	// 将余额字符串转换为decimal类型
	value, err := decimal.NewFromString(resp.Value.Amount)
	if err != nil {
		return TokenAccount{}, fmt.Errorf("无法将代币余额转换为decimal类型: %w", err)
	}

	return TokenAccount{
		Amount:   value,
		Decimals: resp.Value.Decimals,
	}, nil
}

// Close closes the client.
func (e client) Close() error {
	if e.clientRPC != nil {
		return e.clientRPC.Close()
	}

	return nil
}
