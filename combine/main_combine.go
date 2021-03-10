package main

import (
	"context"
	"fmt"
	"flag"
	"math/rand"
	"reflect"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
)

var (
	inputFile = flag.String("input_file", "", "Input file.")
	outputFile = flag.String("output_file", "", "Output file.")
	logN = flag.Uint64("log_n", 8, "Vector size bit.")
	combineDomain = flag.Uint64("combine_domain", 100000, "Number of keys for pre-combine.")
)

func init() {
	beam.RegisterType(reflect.TypeOf((*addRandomKeyFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*genVecFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*combineVecFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*flattenVecFn)(nil)).Elem())
}

type pairedVec struct {
	Vec1 []uint64
	Vec2 []uint64
}

type genVecFn struct {
	LogN uint64
	vecCounter beam.Counter
}

func (fn *genVecFn) Setup() {
	fn.vecCounter = beam.NewCounter("combinetest","genVecFn-vec-count")
}

func (fn *genVecFn) ProcessElement(ctx context.Context, l string, emit func(pairedVec)) {
	fn.vecCounter.Inc(ctx, 1)

	len := 1 << fn.LogN
	e := pairedVec{}
	e.Vec1 = make([]uint64, len)
	e.Vec2 = make([]uint64, len)
	for i := 0; i < len; i++ {
		e.Vec1[i] = rand.Uint64()
		e.Vec2[i] = rand.Uint64()
	}
	emit(e)
}

type combineVecFn struct {
	LogN uint64
	counterInput beam.Counter
	counterCreate beam.Counter
	counterMerge beam.Counter
}

func (fn *combineVecFn) Setup() {
	fn.counterInput = beam.NewCounter("combinetest","combine_input")
	fn.counterCreate = beam.NewCounter("combinetest","combine_create")
	fn.counterMerge = beam.NewCounter("combinetest","combine_merge")
}

type addRandomKeyFn struct {
	Domain uint64
}

func (fn *addRandomKeyFn) ProcessElement(e pairedVec) (uint64, pairedVec) {
	key := uint64(rand.Int63n(int64(fn.Domain)))
	return key, e
}


func (fn *combineVecFn) CreateAccumulator(ctx context.Context) pairedVec {
	fn.counterCreate.Inc(ctx, 1)

	len := 1 << fn.LogN
	return pairedVec{Vec1: make([]uint64, len), Vec2: make([]uint64, len)}
}

func (fn *combineVecFn) AddInput(ctx context.Context, e , p pairedVec) pairedVec {
	fn.counterInput.Inc(ctx, 1)

	len := 1 << fn.LogN
	for i := 0; i < len; i++ {
		e.Vec1[i] += p.Vec1[i]
		e.Vec2[i] += p.Vec2[i]
	}
	return e
}

func (fn *combineVecFn) MergeAccumulators(ctx context.Context, a, b pairedVec) pairedVec {
	fn.counterMerge.Inc(ctx, 1)

	len := 1 << fn.LogN
	for i := 0; i < len; i++ {
		a.Vec1[i] += b.Vec1[i]
		a.Vec2[i] += a.Vec2[i]
	}
	return a
}

type flattenVecFn struct {
	counterInput beam.Counter
	counterOutput beam.Counter
}

func (fn *flattenVecFn) Setup(ctx context.Context) {
	fn.counterInput = beam.NewCounter("combinetest","flattenVecFn_input_count")
	fn.counterOutput = beam.NewCounter("combinetest","flattenVecFn_output_count")
}

func (fn *flattenVecFn) ProcessElement(ctx context.Context, vec pairedVec, emit func(string)) error {
	fn.counterInput.Inc(ctx, 1)

	l := len(vec.Vec1)
	for i := 0; i < l; i++ {
		fn.counterOutput.Inc(ctx, 1)

		emit(fmt.Sprintf("%d,%d,%d", i, vec.Vec1[i], vec.Vec2[i]))
	}
	return nil
}

func main() {
	flag.Parse()

	beam.Init()
	ctx := context.Background()
	pipeline := beam.NewPipeline()
	scope := pipeline.Root()

	records := textio.ReadSdf(scope, *inputFile)
	rRecords := beam.Reshuffle(scope, records)

	vecs := beam.ParDo(scope, &genVecFn{LogN: *logN}, rRecords)

	keyVecs := beam.ParDo(scope, &addRandomKeyFn{Domain: *combineDomain}, vecs)
	combinedKeyVecs := beam.CombinePerKey(scope, &combineVecFn{LogN: *logN}, keyVecs)
	combinedVecs := beam.DropKey(scope, combinedKeyVecs)

	histogram := beam.Combine(scope, &combineVecFn{LogN: *logN}, combinedVecs)

	lines := beam.ParDo(scope, &flattenVecFn{}, histogram)
	textio.Write(scope, *outputFile, lines)

	if err := beamx.Run(ctx, pipeline); err != nil {
		log.Exitf(ctx, "Failed to execute job: %s", err)
	}
}
