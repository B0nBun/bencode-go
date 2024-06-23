Example usage:

```go
type Foo struct {
    Name string `bencode:"name"`
    Value int `bencode:"value"`
    OptionalList []int `bencode:"?list"` // '?' means that the key is optional
}

func main() {
    foo := Foo{}
    err := bencode.Unmarshal([]byte("d4:name6:b0nbun5:valuei999e4:listli1ei2ei3eee"), &foo)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(foo) // {b0nbun 999 [1 2 3]}
}
```