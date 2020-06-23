package models

type Mapper interface {
	MapObjectToMgo(obj interface{}) (interface{}, error)
	MapMgoToObject(obj interface{}) (interface{}, error)
}
