package model

type CollectionDTO struct {
	Session *SessionDTO `json:"session"`
	Events  []EventDTO  `json:"events"`
	Traces  []TraceDTO  `json:"traces"`
}
