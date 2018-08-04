package main

import (
	"time"
)

//-------------------------------------------------------------------------------------------------
type LongRunningRecognizeRequest struct {
	Signal					[]byte					`json:"signal"`
	LanguageCode			string					`json:"language_code"`
	ExecuteBeamSearch		bool					`json:"execute_beam_search"`
}

//-------------------------------------------------------------------------------------------------
type LongRunningRecognizeResponse struct {
	OperationId				uint64					`json:"operation_id"`
}

//-------------------------------------------------------------------------------------------------
type FetchOperationResponse struct {
	Id						uint64					`json:"id"`
	LanguageCode			string					`json:"language_code"`
	BeamSearch				bool					`json:"beam_search"`
	ProcessingStatus		string					`json:"processing_status"`
	ProcessingStartedAt		time.Time				`json:"processing_started_at"`
	ProcessingFinishedAt	time.Time				`json:"processing_finished_at"`
	Speakers				[]*Speaker				`json:"speakers"`
	Transcriptions			[]*Transcription		`json:"transcriptions"`
}

//-------------------------------------------------------------------------------------------------
type Speaker struct {
	Id						uint32					`json:"id"`
	Gender					string					`json:"gender"`
}

//-------------------------------------------------------------------------------------------------
type Alternative struct {
	Transcript				string					`json:"transcript"`
	Confidence				float32					`json:"confidence"`
}

//-------------------------------------------------------------------------------------------------
type Transcription struct {
	TimeStart				float32					`json:"time_start"`
	TimeEnd					float32					`json:"time_end"`
	SpeakerId				uint32					`json:"speaker_id"`
	Alternatives			[]Alternative			`json:"alternatives"`
}
