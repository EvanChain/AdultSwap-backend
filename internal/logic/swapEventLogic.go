package logic

import (
	"awesomeProject3/internal/global"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// SwapEvent Swap 事件结构（包含价格数据）
type SwapEvent struct {
	TxHash      common.Hash
	BlockNumber uint64
	Timestamp   time.Time
	Sender      common.Address
	Recipient   common.Address
	AmountETH   float64 // ETH 数量（带符号）
	AmountUSDC  float64 // USDC 数量（带符号）
	Price       float64 // 计算后的实际价格（USDC per ETH）
}

func GetSwapEventInfo() string {
	// 连接以太坊节点
	client, err := ethclient.Dial(global.BlockChainConfig.RpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	// ETH-USDC 0.3% 池地址
	poolAddress := common.HexToAddress(global.BlockChainConfig.ETHToUSDCAddress)

	// 获取最新区块
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	endBlock := header.Number.Uint64()

	// 计算30分钟前的区块
	startBlock, err := getBlockNumber30MinutesAgo(client, endBlock)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("查询区块范围: %d (大约30分钟前) 到 %d (最新)", startBlock, endBlock)

	// 创建过滤查询
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(endBlock)),
		Addresses: []common.Address{poolAddress},
		Topics: [][]common.Hash{
			{common.HexToHash(global.BlockChainConfig.SwapHash)}, // Swap 事件签名
		},
	}

	// 获取日志
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	// 解析日志
	var events []SwapEvent
	for _, vLog := range logs {
		event, err := parseSwapEvent(client, vLog)
		if err != nil {
			log.Printf("解析事件失败: %v", err)
			continue
		}
		events = append(events, *event)
	}

	// 打印结果
	return printResults(events)
}

// 获取大约30分钟前的区块号
func getBlockNumber30MinutesAgo(client *ethclient.Client, latestBlock uint64) (uint64, error) {
	// 以太坊平均出块时间约为12秒
	blocksPerMinute := 5 // 60/12
	targetBlocks := uint64(60 * blocksPerMinute)

	if latestBlock > targetBlocks {
		return latestBlock - targetBlocks, nil
	}
	return 0, nil
}

// 解析 Swap 事件
func parseSwapEvent(client *ethclient.Client, log types.Log) (*SwapEvent, error) {
	if len(log.Topics) < 3 || len(log.Data) < 160 {
		return nil, fmt.Errorf("无效的日志数据")
	}

	// 获取区块时间戳
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		return nil, fmt.Errorf("获取区块失败: %v", err)
	}

	amount0 := new(big.Int).SetBytes(log.Data[0:32])
	amount1 := new(big.Int).SetBytes(log.Data[32:64])
	sqrtPriceX96 := new(big.Int).SetBytes(log.Data[64:96])

	// 假设token0是WETH，token1是USDC
	ethAmount := formatTokenValue(amount0, 18)
	usdcAmount := formatTokenValue(amount1, 6)
	price := calculatePriceFromSqrtX96(sqrtPriceX96)

	return &SwapEvent{
		TxHash:      log.TxHash,
		BlockNumber: log.BlockNumber,
		Timestamp:   time.Unix(int64(block.Time()), 0),
		Sender:      common.BytesToAddress(log.Topics[1].Bytes()),
		Recipient:   common.BytesToAddress(log.Topics[2].Bytes()),
		AmountETH:   ethAmount,
		AmountUSDC:  usdcAmount,
		Price:       price,
	}, nil
}

// 从 sqrtPriceX96 计算实际价格（USDC per ETH）
func calculatePriceFromSqrtX96(sqrtPriceX96 *big.Int) float64 {
	sqrtPrice := new(big.Float).SetInt(sqrtPriceX96)
	two96 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil))

	priceRatio := new(big.Float).Quo(sqrtPrice, two96)
	priceRatio.Mul(priceRatio, priceRatio)

	// 调整精度 (ETH 18d, USDC 6d)
	priceRatio.Mul(priceRatio, big.NewFloat(1e12))
	price, _ := priceRatio.Float64()
	return price
}

// 格式化代币数量
func formatTokenValue(amount *big.Int, decimals int) float64 {
	value := new(big.Float).SetInt(amount)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	result, _ := new(big.Float).Quo(value, divisor).Float64()

	if amount.Sign() < 0 {
		return -result
	}
	return result
}

// 打印结果
func printResults(events []SwapEvent) string {
	var question strings.Builder
	question.WriteString("\n=== uniswapV3最近30分钟内USDT-ETH交易分析 ===\n")
	question.WriteString("共发现 ")
	question.WriteString(strconv.Itoa(len(events)))
	question.WriteString(" 笔交易\n")
	fmt.Printf("\n=== uniswapV3最近30分钟内USDT-ETH交易分析 ===\n")
	fmt.Printf("共发现 %d 笔交易\n", len(events))

	for i, event := range events {
		question.WriteString("\n交易 #")
		question.WriteString(strconv.Itoa(i + 1))
		question.WriteString("\n")
		question.WriteString("时间: ")
		question.WriteString(event.Timestamp.Format("2006-01-02 15:04:05"))
		question.WriteString("\n")
		question.WriteString("交易哈希: ")
		question.WriteString(event.TxHash.Hex())
		question.WriteString("\n")
		question.WriteString("区块: ")
		question.WriteString(strconv.FormatUint(event.BlockNumber, 10))
		question.WriteString("\n")
		question.WriteString("ETH 数量: ")
		question.WriteString(fmt.Sprintf("%.8f", event.AmountETH))
		question.WriteString("\n")
		question.WriteString("USDC 数量: ")
		question.WriteString(fmt.Sprintf("%.2f", event.AmountUSDC))
		question.WriteString("\n")
		question.WriteString("价格: 1 ETH =")
		question.WriteString(fmt.Sprintf("%.2f", event.Price))
		question.WriteString("USDC\n")
		fmt.Printf("\n交易 #%d\n", i+1)
		fmt.Printf("时间: %s\n", event.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("交易哈希: %s\n", event.TxHash.Hex())
		fmt.Printf("区块: %d\n", event.BlockNumber)
		fmt.Printf("ETH 数量: %.8f\n", event.AmountETH)
		fmt.Printf("USDC 数量: %.2f\n", event.AmountUSDC)
		fmt.Printf("价格: 1 ETH = %.2f USDC\n", event.Price)
	}

	if len(events) > 0 {
		avgPrice := calculateAveragePrice(events)
		fmt.Printf("\n平均价格: 1 ETH = %.2f USDC\n", avgPrice)
	}
	//调用ai
	mistral, err := PostMistral(question.String())
	if err != nil {
		return ""
	}
	return mistral
}

// 计算平均价格
func calculateAveragePrice(events []SwapEvent) float64 {
	total := 0.0
	for _, event := range events {
		total += event.Price
	}
	return total / float64(len(events))
}
