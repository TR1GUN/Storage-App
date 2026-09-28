package schemas

// Record — запись хранилища.
// idx — уникальный идентификатор (задаётся пользователем).
// record — данные в свободном формате.
type Record struct {
	ID   int `json:"idx"`
	Data any `json:"record"`
}
