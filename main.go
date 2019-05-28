package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

const defaultAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ1234567890_:!?+-"

var db *bolt.DB

type Val struct {
	Value string `json:"url"`
	Mode  int    `json:"mode"`
}

type Auth struct {
	Key string `json:"key"`
}

type KeyValue struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Delete string `json:"delete"`
}

func main() {
	var err error
	rand.Seed(time.Now().UnixNano())
	db, err = bolt.Open("bolt.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("buck"))
		if err != nil {
			return fmt.Errorf("Fehler beim Bucket erstellen")
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	r := gin.Default()
	r.PUT("/", func(c *gin.Context) {
		r := Val{}
		c.BindJSON(&r)
		fmt.Printf("Die Url '%s' soll mit dem Modus %d gespeichert werden.\n", r.Value, r.Mode)
		if r.Value == "" {
			c.String(406, "Es kann keine leere URL gespeichert werden.")
		} else {
			key := put(r.Value, r.Mode)
			c.JSON(201, gin.H{
				"key": key,
			})
		}
	})

	r.GET("/:key", func(c *gin.Context) {
		key := c.Param("key")
		if strings.HasPrefix(key, "admin") {
			c.File("admin.html")
		} else if strings.HasPrefix(key, "favicon.ico") {
			c.AbortWithStatus(404)
		} else if strings.HasPrefix(key, "main.css") {
			c.File("main.css")
		} else {
			value := get(key)
			if value != "" {
				if strings.HasPrefix(value, "https://") {
					c.Redirect(302, value)
				} else if strings.HasPrefix(value, "http://") {
					c.Redirect(302, value)
				} else {
					c.Redirect(302, "http://"+value)
				}
			} else {
				c.Redirect(303, "/?key="+key)
			}
		}

	})

	r.DELETE("/:key", func(c *gin.Context) {
		key := c.Param("key")
		value := get(key)
		if value != "" {
			remove(key, value)
			c.JSON(200, gin.H{
				"deleted": true,
				"key":     key,
				"value":   value,
			})
		} else {
			c.String(404, "I'm sorry, but the key '%s' is not linked to something.", key)
		}
	}) //TODO: Require Secrete

	r.POST("/admin", func(c *gin.Context) {
		k := Auth{}
		c.BindJSON(&k)
		fmt.Println(k.Key)
		if k.Key == "test" {
			all := getAll()
			c.JSON(200, gin.H{
				"list": all,
			})
		} else {
			c.AbortWithStatus(401)
		}
	})

	r.StaticFile("/", "index.html")
	r.Run(":2407")
}

func get(key string) string {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket([]byte("buck")).Get([]byte(key))
		fmt.Printf("Gelesen! Key: '%s', Value: '%s'\n", key, string(value))
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return ""
	}
	return string(value)
}

func put(value string, mode int) string {
	key := keygen(mode)
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("buck")).Put([]byte(key), []byte(value))
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return ""
	}
	fmt.Printf("Gespeichert! Key: '%s', Value: '%s'\n", key, value)
	return key
}

func remove(key string, value string) {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("buck")).Delete([]byte(key))
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	fmt.Printf("Gelöscht! Key: '%s', Value: '%s'\n", key, value)
}

func keygen(mode int) string {
	var nc int
	switch mode {
	case 1:
		nc = 4
	case 2:
		nc = 8
	case 3:
		nc = 12
	case 4:
		nc = 20
	case 5:
		nc = 40
	default:
		nc = 8
	}
	var key strings.Builder
	for i := 0; i < nc; i++ {
		key.WriteRune([]rune(defaultAlphabet)[rand.Intn(len(defaultAlphabet))])
	}
	if get(key.String()) != "" || key.String() == "admin" {
		return keygen(mode)
	}
	return key.String()
}

func getAll() []KeyValue {
	var res []KeyValue
	err := db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("buck"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			tmp := KeyValue{}
			tmp.Key = string(k)
			tmp.Value = string(v)
			tmp.Delete = "<input type='button' onclick='deletee(\"" + string(k) + "\")' value='Löschen'>"
			res = append(res, tmp)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
	return res
}
