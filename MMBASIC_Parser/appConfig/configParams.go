package appConfig

// ***************************************************************************************************
// * Update configuration parameters here                                                            *
// * Parameter name must start with uppercase letter to be exportable to other packages              *
// * Remember to set constant (default) value below too, and update Conf var declaration accordingly *
// ***************************************************************************************************
type Config struct {
	Stringify  bool   `json:"Data_is_text"` // Data is text and therefore needs to be enclosed in quotation marks
	ColCount   int    `json:"Column_Count"` // Maximum length of each Data line
	Delimiter  string `json:"Delimiter"`    // Delimiter
	SourceFile string `json:"Source_File"`  // Source File path and name
	TargetFile string `json:"Target_File"`  // Target File path and name
}

// Default calibration and configuration values for the sensor array
const (
	Stringify  bool   = true             // Data treated as text
	ColCount   int    = 80               // 80 Columns
	Delimiter  string = ","              // Comma Separated Values
	SourceFile string = "DataSource.csv" // Sourc file name
	TargetFile string = "Output.bas"
)

var Conf = Config{
	Stringify,
	ColCount,
	Delimiter,
	SourceFile,
	TargetFile,
}

//****************************************************************************************************

const (
	configFileName string = "data_parser.conf"
)
