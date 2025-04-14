package model

type CollectionDTO struct {
	Session *SessionDTO `json:"session" validation:"omitnil,required"`
	Events  []EventDTO  `json:"events" validation:"gt=0,dive,required"`
	Traces  []TraceDTO  `json:"traces" validation:"dive,required"`
}
