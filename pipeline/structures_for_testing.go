package pipeline

// this structures needs for writing unit tests in other's issues
type (
	testStructWithSimpleTypes struct {
		Field1 string  `cf_hocon:"field1"`
		Field2 int     `cf_hocon:"field2"`
		Field3 int8    `cf_hocon:"field3"`
		Field4 int16   `cf_hocon:"field4"`
		Field5 int32   `cf_hocon:"field5"`
		Field6 int64   `cf_hocon:"field6"`
		Field7 float32 `cf_hocon:"field7"`
		Field8 float64 `cf_hocon:"field8"`
		Field9 bool    `cf_hocon:"field9"`
	}

	// testStructWithComplexSlices struct {
	// 	Field10 []string      `cf_hocon:"field10"`
	// 	Field11 []int         `cf_hocon:"field11"`
	// 	Field12 []int8        `cf_hocon:"field12"`
	// 	Field13 []int16       `cf_hocon:"field13"`
	// 	Field14 []int32       `cf_hocon:"field14"`
	// 	Field15 []int64       `cf_hocon:"field15"`
	// 	Field16 []float32     `cf_hocon:"field16"`
	// 	Field17 []float64     `cf_hocon:"field17"`
	// 	Field18 []bool        `cf_hocon:"field18"`
	// 	Field19 []interface{} `cf_hocon:"field19"`
	// }

	// testStructWithComplextMaps struct {
	// 	Field18 map[string]string                      `cf_hocon:"field18"`
	// 	Field19 map[string]int                         `cf_hocon:"field19"`
	// 	Field20 map[string]int8                        `cf_hocon:"field20"`
	// 	Field21 map[string]int16                       `cf_hocon:"field21"`
	// 	Field22 map[string]int32                       `cf_hocon:"field22"`
	// 	Field23 map[string]int64                       `cf_hocon:"field23"`
	// 	Field24 map[string]float32                     `cf_hocon:"field24"`
	// 	Field25 map[string]float64                     `cf_hocon:"field25"`
	// 	Field26 map[string]bool                        `cf_hocon:"field26"`
	// 	Field27 map[string]interface{}                 `cf_hocon:"field27"`
	// 	Field28 map[string]testStructWithSimpleTypes   `cf_hocon:"field28"`
	// 	Field17 map[string][]string                    `cf_hocon:"field17"`
	// 	Field29 map[string][]int                       `cf_hocon:"field29"`
	// 	Field30 map[string][]int8                      `cf_hocon:"field30"`
	// 	Field31 map[string][]int16                     `cf_hocon:"field31"`
	// 	Field32 map[string][]int32                     `cf_hocon:"field32"`
	// 	Field33 map[string][]int64                     `cf_hocon:"field33"`
	// 	Field16 map[string][]testStructWithSimpleTypes `cf_hocon:"field16"`
	// 	Field15 map[string][]interface{}               `cf_hocon:"field15"`
	// 	Field34 map[int]string                         `cf_hocon:"field34"`
	// 	Field35 map[int]int                            `cf_hocon:"field35"`
	// 	Field36 map[int]int8                           `cf_hocon:"field36"`
	// 	Field37 map[int]int16                          `cf_hocon:"field37"`
	// 	Field38 map[int]int32                          `cf_hocon:"field38"`
	// 	Field39 map[int]int64                          `cf_hocon:"field39"`
	// 	Field40 map[int]float32                        `cf_hocon:"field40"`
	// 	Field41 map[int]float64                        `cf_hocon:"field41"`
	// 	Field42 map[int]bool                           `cf_hocon:"field42"`
	// 	Field43 map[int]interface{}                    `cf_hocon:"field43"`
	// 	Field44 map[int][]string                       `cf_hocon:"field44"`
	// 	Field45 map[int][]int                          `cf_hocon:"field45"`
	// 	Field46 map[int][]int8                         `cf_hocon:"field46"`
	// 	Field47 map[int][]int16                        `cf_hocon:"field47"`
	// 	Field48 map[int][]int32                        `cf_hocon:"field48"`
	// 	Field49 map[int][]int64                        `cf_hocon:"field49"`
	// 	Field50 map[int][]testStructWithSimpleTypes    `cf_hocon:"field50"`
	// 	Field51 map[int][]interface{}                  `cf_hocon:"field51"`
	// }

	// testStructWithNestedStruct struct {
	// 	Field1 testStructWithSimpleTypes   `cf_hocon:"node=simpler,path=mystuct`
	// 	Field2 testStructWithComplexSlices `cf_hocon:"node=complexer,path=myStructSlices"`
	// 	Field3 testStructWithComplextMaps  `cf_hocon:"node=complexer,paht=myStructMaps"`
	// }
)
