package schemas

// Record — полный аналог pydantic-модели:
//
//	class Record(BaseMode
//	class Record(BaseModel):
//	    """Наша ожидаемая запись."""
//	    idx: int
//	    record: Any
type Record struct {
	Idx    int `json:"idx"`
	Record Map `json:"record"`
}
