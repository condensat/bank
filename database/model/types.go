package model

type ID uint64
type RefID ID

type String string
type Float float64

type Base58 String
type ZeroInt *int
type ZeroFloat *Float

type Model interface{}
