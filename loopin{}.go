 
package main
import "fmt"

type album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}
 

func main() {
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
    {ID: "4", Title: "Sarah Vaughan and Clifford new user ", Artist: "Sarah Vaughan", Price: 11.99},

for _ ,num :=range albums {
    fmt.Println(num)
}

}
