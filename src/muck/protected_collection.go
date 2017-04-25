package fbmuck

type MLevel int

type ProtectedCollection interface {
	GetAs(int, interface{})
	SetAs(int, interface{}, interface{})
}