package objects

import (
	"math/rand"
)

type BaseObject interface {
	Type() string
}

type ObjectOne struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	IsHappy bool   `json:"is_happy"`
	Comment string `json:"comment"`
}

func (o *ObjectOne) Type() string {
	return "object.one"
}

type ObjectTwo struct {
	Id    int64    `json:"id"`
	Mode  string   `json:"mode"`
	Items []string `json:"items"`
}

func (o *ObjectTwo) Type() string {
	return "object.two"
}

type ObjectThree struct {
	Name       string `json:"name"`
	Rating     uint64 `json:"rating"`
	RatingType string `json:"type"`
}

func (o *ObjectThree) Type() string {
	return "object.three"
}

var Choices = []string{"aaaa", "bbbb", "cccc", "dddd"}
var RatingChoices = []string{"happy", "sad", "sorry", "goodbye"}

const letterBytes = "0123456789abcdefghijklmnopqrs tuv wxy\n\t\rzABCDEFGHIJKLMNOPQ RSTUVWXYZ"

func randText(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandomObject() BaseObject {
	gots := rand.Int63n(3)
	switch gots {
	case 1:
		items := []string{}
		for i := 0; i < rand.Intn(10); i++ {
			items = append(items, randText(rand.Intn(50)))
		}
		return &ObjectTwo{
			Id:    rand.Int63(),
			Mode:  Choices[rand.Intn(len(Choices))],
			Items: items,
		}
	case 2:
		return &ObjectThree{
			Name:       randText(32),
			Rating:     uint64(rand.Int63n(5)),
			RatingType: RatingChoices[rand.Intn(len(RatingChoices))],
		}
	default:
		return &ObjectOne{
			Id:      rand.Int63(),
			Name:    randText(32),
			IsHappy: rand.Intn(2) > 1,
			Comment: randText(1024),
		}
	}
}
