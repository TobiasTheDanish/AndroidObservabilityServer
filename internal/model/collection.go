package model

type CollectionDTO struct {
	Session *SessionDTO `json:"session" validation:"omitnil,required"`
	Events  []EventDTO  `json:"events"`
	Traces  []TraceDTO  `json:"traces"`
}
