package faiss

import (
	"context"
	"infer-microservices/common"
	faiss_index "infer-microservices/common/faiss_gogofaster"
	"infer-microservices/common/flags"
	"infer-microservices/utils"
	"time"
)

var FaissIndexClientInstance *FaissIndexClient
var grpcTimeout int64

type FaissIndexClient struct {
	indexName     string                     //index name.
	faissGrpcPool *common.GRPCPool           //faiss  grpc pool.
	faissIndexs   *faiss_index.RecallRequest // faiss index.
}

// INFO: singleton instance
func init() {
	flagFactory := flags.FlagFactory{}
	flagTensorflow := flagFactory.FlagTensorflowFactory()
	grpcTimeout = *flagTensorflow.GetTfservingTimeoutMs()
}

// index name
func (f *FaissIndexClient) setIndexName(indexName string) {
	f.indexName = indexName
}

func (f *FaissIndexClient) GetIndexName() string {
	return f.indexName
}

// grpc pool
func (f *FaissIndexClient) setFaissGrpcPool(faissGrpcPool *common.GRPCPool) {
	f.faissGrpcPool = faissGrpcPool
}

func (f *FaissIndexClient) GetFaissGrpcPool() *common.GRPCPool {
	return f.faissGrpcPool
}

// FaissIndexs
func (f *FaissIndexClient) setFaissIndexs(faissIndexs *faiss_index.RecallRequest) {
	f.faissIndexs = faissIndexs
}

func (f *FaissIndexClient) GetFaissIndexs() *faiss_index.RecallRequest {
	return f.faissIndexs
}

// faiss index conf load
func (f *FaissIndexClient) ConfigLoad(domain string, dataId string, indexConfStr string) error {
	dataConf := utils.Json2Map(indexConfStr)

	// create faiss grpc pool
	faissGrpcConf := dataConf["faissGrpcAddr"].(map[string]interface{})
	faissGrpcPool, err := common.CreateGrpcConn(faissGrpcConf)
	if err != nil {
		return err
	}

	indexInfo := dataConf["indexInfo"].(map[string]interface{})
	for indexName, tmpIndexConf := range indexInfo { //only 1 index
		tmpIndexConfMap := tmpIndexConf.(map[string]interface{})
		recallNum := int32(tmpIndexConfMap["recall_num"].(float64))
		signature := tmpIndexConfMap["index_name"].(string)
		indexInfoStruct := &faiss_index.RecallRequest{
			IndexName: signature,
			RecallNum: recallNum,
		}

		f.setIndexName(indexName)
		f.setFaissGrpcPool(faissGrpcPool)
		f.setFaissIndexs(indexInfoStruct)
	}

	return nil
}

func (f *FaissIndexClient) FaissVectorSearch(example common.ExampleFeatures, vector []float32) ([]*faiss_index.ItemInfo, error) {
	faissIndexs := f.GetFaissIndexs()
	faissGrpcConn, err := f.GetFaissGrpcPool().Get()
	if err != nil {
		return nil, err
	}

	defer f.GetFaissGrpcPool().Put(faissGrpcConn)

	faissClient := faiss_index.NewGrpcRecallServerServiceClient(faissGrpcConn)
	vector_info := faiss_index.UserVectorInfo{
		UserVector: vector,
	}

	index_conf_tmp := &faiss_index.RecallRequest{
		IndexName:       faissIndexs.IndexName,
		UserVectorInfo_: &vector_info,
		RecallNum:       faissIndexs.RecallNum,
	}

	if len(*example.UserExampleFeatures.Buff) == 0 || len(*example.UserContextExampleFeatures.Buff) == 0 {
		return make([]*faiss_index.ItemInfo, 0), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(grpcTimeout)*time.Millisecond)
	defer cancel()

	rst, err := faissClient.GrpcRecall(ctx, index_conf_tmp)
	if err != nil {
		return nil, err
	}

	return rst.ItemInfo_, nil
}
