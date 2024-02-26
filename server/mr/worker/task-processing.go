package worker

import (
	"log"
	"os"
	"path/filepath"
	"server/mr/common"
	"strconv"
)

// processTask dispatches the task to either map or reduce processing.
func processTask(reply RequestTaskReply) {
	if reply.MapJob != nil {
		processMapTask(reply.MapJob)
	} else if reply.ReduceJob != nil {
		processReduceTask(reply.ReduceJob)
	}
}

// processMapTask handles the map task, including reading input, executing the map function, and storing the output.
func processMapTask(job *MapJob) {

	// call common.ConvertJsonToCairo(job.InputFile) -> outputs cairo data file
	// if no errors,
	// call common.CallCairoMap() -> runs cairo mapper
	//    also handles Cairo shell -> intermediate.json
	// skip partitioning for now
	// return data to coordinator

	_, err := os.ReadFile(job.InputFile)
	if err != nil {
		log.Fatalf("cannot read %v", job.InputFile)
	}

	projectRoot := common.GetProjectRoot()

	log.Printf("DIRECTORY")
	log.Printf(projectRoot)

	jsonDst := filepath.Join(projectRoot, "cairo/map/src/matvecdata_mapper.cairo")
	common.ConvertJsonToCairo(job.InputFile, jsonDst)
	// should probably check if the cairo was written successfully

	// *********** @Trevor Uncomment SECTION TO INJECT CAIRON all ALL INTO ONE FOR TRACES ***********
	aggMapDst := filepath.Join(projectRoot, "cairo/map/src/agg-lib.cairo")
	common.AggregateMapperCairo(aggMapDst)	

	// Call Cairo Map
	mapDst := filepath.Join(projectRoot, "server/data/mr-tmp")
	intermediateFiles := common.CallCairoMap(job.MapJobNumber, mapDst)

	// skip partitioning for now, here's normal way:
	// kva := mapf(job.InputFile, string(content))
	// sort.Sort(ByKey(kva))

	// partitionedKva := partitionByKey(kva, job.ReducerCount)
	// intermediateFiles := writeIntermediateFiles(partitionedKva, job.MapJobNumber)
	reportMapTaskToCoordinator(job.InputFile, intermediateFiles)
}

// processReduceTask handles the reduce task, including reading intermediate files, executing the reduce function, and writing the output.
func processReduceTask(job *ReduceJob) {
	projectRoot := common.GetProjectRoot()

	// TODO:
	// call function to read intermediate file to Cairo
	// TEMP: just 1 reducer for now
	dst := filepath.Join(projectRoot, "cairo/reducer/src/matvecdata_reducer.cairo")
	common.ConvertIntermediateToCairo(job.IntermediateFiles[0], dst)

	// *********** @Trevor Uncomment: SECTION TO INJECT CAIRON all ALL INTO ONE FOR TRACES ***********
	aggRedDst := filepath.Join(projectRoot, "cairo/reducer/src/agg-lib.cairo")
	common.AggregateReducerCairo(aggRedDst)

	reduceDst := filepath.Join(projectRoot, "server/data/mr-tmp")
	reduceNumStr := strconv.Itoa(job.ReduceNumber)
	common.CallCairoReduce(reduceNumStr, reduceDst)

	// intermediate := readIntermediateFiles(job.IntermediateFiles)
	// sort.Sort(ByKey(intermediate))

	// writeReduceOutput(intermediate, job.ReduceNumber, reducef)
	reportReduceTaskToCoordinator(job.ReduceNumber)
}
