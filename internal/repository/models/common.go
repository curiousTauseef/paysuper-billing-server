package models

type MgoMultiLang struct {
	Lang  string `bson:"lang"`
	Value string `bson:"value"`
}

