# gongular2

Sorry for the lame name. It is an idea only repo right now to keep track of my thoughts.

```go
type SomeRequest struct {
	db *sqlx.DB
	es *elastic.Client
	body struct {
		Name string `validate:"notempty"`
		Age  int `validate:"notzero"`
	}
	param struct {
		ID string
	}
	query struct {
		Offset int
	}
	response struct {
		Status int
		Age    float64
	}
}

func (s *SomeRequest) GET(r *http.Request, wr http.ResponseWriter) error {
	// do some stuff and set response of s to something 
}

func (s *SomeRequest) Validate(r *http.Request, wr http.ResponseWriter) error {
	// Custom validation logic
}

router.GET("/path/:ID", SomeReuest{})
router.WS("/ws/:UserName", WebSocketHandler{})
```
