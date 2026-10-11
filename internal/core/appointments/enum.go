package appointments

import (
	"clinic-vet-api/internal/shared/errors"
	"context"
	"fmt"
	"strings"
)

// VisitReason represents the reason for a veterinary visit
type VisitReason string

const (
	VisitReasonRoutineCheckup VisitReason = "routine_checkup"
	VisitReasonVaccination    VisitReason = "vaccination"
	VisitReasonInjury         VisitReason = "injury"
	VisitReasonIllness        VisitReason = "illness"
	VisitReasonDental         VisitReason = "dental"
	VisitReasonGrooming       VisitReason = "grooming"
	VisitReasonBehavior       VisitReason = "behavior"
	VisitReasonNutrition      VisitReason = "nutrition"
	VisitReasonFollowUp       VisitReason = "follow_up"
	VisitReasonEmergency      VisitReason = "emergency"
)

// VisitReason constants and methods
var (
	ValidVisitReasons = []VisitReason{
		VisitReasonRoutineCheckup,
		VisitReasonVaccination,
		VisitReasonInjury,
		VisitReasonIllness,
		VisitReasonDental,
		VisitReasonGrooming,
		VisitReasonBehavior,
		VisitReasonNutrition,
		VisitReasonFollowUp,
		VisitReasonEmergency,
	}

	visitReasonMap = map[string]VisitReason{
		"routine_checkup": VisitReasonRoutineCheckup,
		"routine checkup": VisitReasonRoutineCheckup,
		"checkup":         VisitReasonRoutineCheckup,
		"vaccination":     VisitReasonVaccination,
		"injury":          VisitReasonInjury,
		"illness":         VisitReasonIllness,
		"dental":          VisitReasonDental,
		"grooming":        VisitReasonGrooming,
		"behavior":        VisitReasonBehavior,
		"nutrition":       VisitReasonNutrition,
		"follow_up":       VisitReasonFollowUp,
		"follow up":       VisitReasonFollowUp,
		"followup":        VisitReasonFollowUp,
		"emergency":       VisitReasonEmergency,
	}

	visitReasonDisplayNames = map[VisitReason]string{
		VisitReasonRoutineCheckup: "Routine Checkup",
		VisitReasonVaccination:    "Vaccination",
		VisitReasonInjury:         "Injury",
		VisitReasonIllness:        "Illness",
		VisitReasonDental:         "Dental Care",
		VisitReasonGrooming:       "Grooming",
		VisitReasonBehavior:       "Behavior Consultation",
		VisitReasonNutrition:      "Nutrition Consultation",
		VisitReasonFollowUp:       "Follow-up Visit",
		VisitReasonEmergency:      "Emergency Visit",
	}
)

func (vr VisitReason) IsValid() bool {
	_, exists := visitReasonMap[string(vr)]
	return exists
}

func ParseVisitReason(reason string) (VisitReason, error) {
	normalized := normalizeInput(reason)
	if val, exists := visitReasonMap[normalized]; exists {
		return val, nil
	}
	return "", InvalidEnumParserError("VisitReason", reason)
}

func MustParseVisitReason(reason string) VisitReason {
	parsed, err := ParseVisitReason(reason)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (vr VisitReason) String() string {
	return string(vr)
}

func (vr VisitReason) DisplayName() string {
	if displayName, exists := visitReasonDisplayNames[vr]; exists {
		return displayName
	}
	return "Unknown Reason"
}

func (vr VisitReason) Values() []VisitReason {
	return ValidVisitReasons
}

// VisitType represents the type of veterinary visit
type VisitType string

const (
	VisitTypePhysicalExam   VisitType = "physical_exam"
	VisitTypeSurgery        VisitType = "surgery"
	VisitTypeEmergencyVisit VisitType = "emergency_visit"
	VisitTypeFollowUp       VisitType = "follow_up"
	VisitTypeConsultation   VisitType = "consultation"
	VisitTypeProcedure      VisitType = "procedure"
	VisitTypeVaccination    VisitType = "vaccination"
	VisitTypeDental         VisitType = "dental"
	VisitTypeGrooming       VisitType = "grooming"
)

// VisitType constants and methods
var (
	ValidVisitTypes = []VisitType{
		VisitTypePhysicalExam,
		VisitTypeSurgery,
		VisitTypeEmergencyVisit,
		VisitTypeFollowUp,
		VisitTypeConsultation,
		VisitTypeProcedure,
		VisitTypeVaccination,
		VisitTypeDental,
		VisitTypeGrooming,
	}

	visitTypeMap = map[string]VisitType{
		"physical_exam":   VisitTypePhysicalExam,
		"physical exam":   VisitTypePhysicalExam,
		"exam":            VisitTypePhysicalExam,
		"surgery":         VisitTypeSurgery,
		"emergency_visit": VisitTypeEmergencyVisit,
		"emergency visit": VisitTypeEmergencyVisit,
		"emergency":       VisitTypeEmergencyVisit,
		"follow_up":       VisitTypeFollowUp,
		"follow up":       VisitTypeFollowUp,
		"followup":        VisitTypeFollowUp,
		"consultation":    VisitTypeConsultation,
		"consult":         VisitTypeConsultation,
		"procedure":       VisitTypeProcedure,
		"vaccination":     VisitTypeVaccination,
		"dental":          VisitTypeDental,
		"grooming":        VisitTypeGrooming,
	}

	visitTypeDisplayNames = map[VisitType]string{
		VisitTypePhysicalExam:   "Physical Examination",
		VisitTypeSurgery:        "Surgery",
		VisitTypeEmergencyVisit: "Emergency Visit",
		VisitTypeFollowUp:       "Follow-up Visit",
		VisitTypeConsultation:   "Consultation",
		VisitTypeProcedure:      "Medical Procedure",
		VisitTypeVaccination:    "Vaccination",
		VisitTypeDental:         "Dental Procedure",
		VisitTypeGrooming:       "Grooming Service",
	}

	visitTypeDurations = map[VisitType]int{
		VisitTypePhysicalExam:   30,
		VisitTypeSurgery:        120,
		VisitTypeEmergencyVisit: 60,
		VisitTypeFollowUp:       20,
		VisitTypeConsultation:   45,
		VisitTypeProcedure:      60,
		VisitTypeVaccination:    15,
		VisitTypeDental:         45,
		VisitTypeGrooming:       60,
	}
)

func (vt VisitType) IsValid() bool {
	_, exists := visitTypeMap[string(vt)]
	return exists
}

func ParseVisitType(visitType string) (VisitType, error) {
	normalized := normalizeInput(visitType)
	if val, exists := visitTypeMap[normalized]; exists {
		return val, nil
	}
	return "", InvalidEnumParserError("VisitType", visitType)
}

func MustParseVisitType(visitType string) VisitType {
	parsed, err := ParseVisitType(visitType)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (vt VisitType) String() string {
	return string(vt)
}

func (vt VisitType) DisplayName() string {
	if displayName, exists := visitTypeDisplayNames[vt]; exists {
		return displayName
	}
	return "Unknown Visit Type"
}

func (vt VisitType) Values() []VisitType {
	return ValidVisitTypes
}

func (vt VisitType) DefaultDuration() int {
	if duration, exists := visitTypeDurations[vt]; exists {
		return duration
	}
	return 30 // default duration in minutes
}

// PetCondition represents the condition of a pet during a visit
type PetCondition string

const (
	PetConditionStable   PetCondition = "stable"
	PetConditionCritical PetCondition = "critical"
	PetConditionFair     PetCondition = "fair"
	PetConditionGood     PetCondition = "good"
	PetConditionSerious  PetCondition = "serious"
	PetConditionGuarded  PetCondition = "guarded"
)

// PetCondition constants and methods
var (
	ValidPetConditions = []PetCondition{
		PetConditionStable,
		PetConditionCritical,
		PetConditionFair,
		PetConditionGood,
		PetConditionSerious,
		PetConditionGuarded,
	}

	petConditionMap = map[string]PetCondition{
		"stable":   PetConditionStable,
		"critical": PetConditionCritical,
		"fair":     PetConditionFair,
		"good":     PetConditionGood,
		"serious":  PetConditionSerious,
		"guarded":  PetConditionGuarded,
	}

	petConditionDisplayNames = map[PetCondition]string{
		PetConditionStable:   "Stable",
		PetConditionCritical: "Critical",
		PetConditionFair:     "Fair",
		PetConditionGood:     "Good",
		PetConditionSerious:  "Serious",
		PetConditionGuarded:  "Guarded",
	}
)

func (pc PetCondition) IsValid() bool {
	_, exists := petConditionMap[string(pc)]
	return exists
}

func ParsePetCondition(condition string) (PetCondition, error) {
	normalized := normalizeInput(condition)
	if val, exists := petConditionMap[normalized]; exists {
		return val, nil
	}
	return "", InvalidEnumParserError("PetCondition", condition)
}

func MustParsePetCondition(condition string) PetCondition {
	parsed, err := ParsePetCondition(condition)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (pc PetCondition) String() string {
	return string(pc)
}

func (pc PetCondition) DisplayName() string {
	if displayName, exists := petConditionDisplayNames[pc]; exists {
		return displayName
	}
	return "Unknown Condition"
}

func (pc PetCondition) Values() []PetCondition {
	return ValidPetConditions
}

// Utility functions
func normalizeInput(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	input = strings.ReplaceAll(input, " ", "_")
	return input
}

func GetAllVisitReasons() []VisitReason {
	return ValidVisitReasons
}

func GetAllVisitTypes() []VisitType {
	return ValidVisitTypes
}

func GetAllPetConditions() []PetCondition {
	return ValidPetConditions
}

func GetUrgentVisitReasons() []VisitReason {
	return []VisitReason{
		VisitReasonEmergency,
		VisitReasonInjury,
		VisitReasonIllness,
	}
}

func GetRoutineVisitReasons() []VisitReason {
	return []VisitReason{
		VisitReasonRoutineCheckup,
		VisitReasonVaccination,
		VisitReasonGrooming,
		VisitReasonDental,
	}
}

func InvalidEnumParserError(enumName, enumValue string) error {
	return errors.InvalidEnumValue(context.Background(), enumValue, enumValue, fmt.Sprintf("invalid %s value: %s", enumName, enumValue), "enum parse")

}

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
	AppointmentStatusPending      AppointmentStatus = "pending"
	AppointmentStatusCancelled    AppointmentStatus = "cancelled"
	AppointmentStatusCompleted    AppointmentStatus = "completed"
	AppointmentStatusRescheduled  AppointmentStatus = "rescheduled"
	AppointmentStatusConfirmed    AppointmentStatus = "confirmed"
	AppointmentStatusNotPresented AppointmentStatus = "not_presented"
)

// AppointmentStatus constants and methods
var (
	ValidAppointmentStatuses = []AppointmentStatus{
		AppointmentStatusPending,
		AppointmentStatusCancelled,
		AppointmentStatusCompleted,
		AppointmentStatusRescheduled,
		AppointmentStatusConfirmed,
		AppointmentStatusNotPresented,
	}

	appointmentStatusMap = map[string]AppointmentStatus{
		"pending":       AppointmentStatusPending,
		"cancelled":     AppointmentStatusCancelled,
		"completed":     AppointmentStatusCompleted,
		"rescheduled":   AppointmentStatusRescheduled,
		"confirmed":     AppointmentStatusConfirmed,
		"not_presented": AppointmentStatusNotPresented,
		"not presented": AppointmentStatusNotPresented,
	}

	appointmentStatusDisplayNames = map[AppointmentStatus]string{
		AppointmentStatusPending:      "Pending",
		AppointmentStatusCancelled:    "Cancelled",
		AppointmentStatusCompleted:    "Completed",
		AppointmentStatusRescheduled:  "Rescheduled",
		AppointmentStatusConfirmed:    "Confirmed",
		AppointmentStatusNotPresented: "Not Presented",
	}
)

func (as AppointmentStatus) IsValid() bool {
	_, exists := appointmentStatusMap[string(as)]
	return exists
}

func ParseAppointmentStatus(status string) (AppointmentStatus, error) {
	normalized := strings.TrimSpace(strings.ToLower(status))
	if val, exists := appointmentStatusMap[normalized]; exists {
		return val, nil
	}

	return "", InvalidEnumParserError("AppointmentStatus", status)
}

func MustParseAppointmentStatus(status string) AppointmentStatus {
	parsed, err := ParseAppointmentStatus(status)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (as AppointmentStatus) String() string {
	return string(as)
}

func (as AppointmentStatus) DisplayName() string {
	if displayName, exists := appointmentStatusDisplayNames[as]; exists {
		return displayName
	}
	return "Unknown Status"
}

func (as AppointmentStatus) Values() []AppointmentStatus {
	return ValidAppointmentStatuses
}

// CanBeCancelled checks if an appointment status allows cancellation
func (as AppointmentStatus) CanBeCancelled() bool {
	return as == AppointmentStatusPending || as == AppointmentStatusConfirmed
}

// ClinicService represents the types of services offered by the clinic
type ClinicService string

const (
	ClinicServiceGeneralConsultation ClinicService = "general_consultation"
	ClinicServiceVaccination         ClinicService = "vaccination"
	ClinicServiceSurgery             ClinicService = "surgery"
	ClinicServiceDentalCare          ClinicService = "dental_care"
	ClinicServiceEmergencyCare       ClinicService = "emergency_care"
	ClinicServiceGrooming            ClinicService = "grooming"
	ClinicServiceNutritionConsult    ClinicService = "nutrition_consult"
	ClinicServiceBehaviorConsult     ClinicService = "behavior_consult"
	ClinicServiceWellnessExam        ClinicService = "wellness_exam"
	ClinicServiceOther               ClinicService = "other"
)

// ClinicService constants and methods
var (
	ValidClinicServices = []ClinicService{
		ClinicServiceGeneralConsultation,
		ClinicServiceVaccination,
		ClinicServiceSurgery,
		ClinicServiceDentalCare,
		ClinicServiceEmergencyCare,
		ClinicServiceGrooming,
		ClinicServiceNutritionConsult,
		ClinicServiceBehaviorConsult,
		ClinicServiceWellnessExam,
		ClinicServiceOther,
	}

	clinicServiceMap = map[string]ClinicService{
		"general_consultation": ClinicServiceGeneralConsultation,
		"general consultation": ClinicServiceGeneralConsultation,
		"consultation":         ClinicServiceGeneralConsultation,
		"vaccination":          ClinicServiceVaccination,
		"surgery":              ClinicServiceSurgery,
		"dental_care":          ClinicServiceDentalCare,
		"dental care":          ClinicServiceDentalCare,
		"dental":               ClinicServiceDentalCare,
		"emergency_care":       ClinicServiceEmergencyCare,
		"emergency care":       ClinicServiceEmergencyCare,
		"emergency":            ClinicServiceEmergencyCare,
		"grooming":             ClinicServiceGrooming,
		"nutrition_consult":    ClinicServiceNutritionConsult,
		"nutrition consult":    ClinicServiceNutritionConsult,
		"nutrition":            ClinicServiceNutritionConsult,
		"behavior_consult":     ClinicServiceBehaviorConsult,
		"behavior consult":     ClinicServiceBehaviorConsult,
		"behavior":             ClinicServiceBehaviorConsult,
		"wellness_exam":        ClinicServiceWellnessExam,
		"wellness exam":        ClinicServiceWellnessExam,
		"wellness":             ClinicServiceWellnessExam,
		"other":                ClinicServiceOther,
	}

	clinicServiceDisplayNames = map[ClinicService]string{
		ClinicServiceGeneralConsultation: "General Consultation",
		ClinicServiceVaccination:         "Vaccination",
		ClinicServiceSurgery:             "Surgery",
		ClinicServiceDentalCare:          "Dental Care",
		ClinicServiceEmergencyCare:       "Emergency Care",
		ClinicServiceGrooming:            "Grooming",
		ClinicServiceNutritionConsult:    "Nutrition Consultation",
		ClinicServiceBehaviorConsult:     "Behavior Consultation",
		ClinicServiceWellnessExam:        "Wellness Exam",
		ClinicServiceOther:               "Other Service",
	}

	clinicServiceDurations = map[ClinicService]int{
		ClinicServiceGeneralConsultation: 30,
		ClinicServiceVaccination:         15,
		ClinicServiceSurgery:             120,
		ClinicServiceDentalCare:          45,
		ClinicServiceEmergencyCare:       60,
		ClinicServiceGrooming:            60,
		ClinicServiceNutritionConsult:    45,
		ClinicServiceBehaviorConsult:     60,
		ClinicServiceWellnessExam:        30,
		ClinicServiceOther:               30,
	}
)

func (cs ClinicService) IsValid() bool {
	_, exists := clinicServiceMap[string(cs)]
	return exists
}

func ParseClinicService(service string) (ClinicService, error) {
	normalized := normalizeServiceInput(service)
	if val, exists := clinicServiceMap[normalized]; exists {
		return val, nil
	}
	return "", InvalidEnumParserError("ClinicService", service)
}

func MustParseClinicService(service string) ClinicService {
	parsed, err := ParseClinicService(service)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (cs ClinicService) String() string {
	return string(cs)
}

func (cs ClinicService) DisplayName() string {
	if displayName, exists := clinicServiceDisplayNames[cs]; exists {
		return displayName
	}
	return "Unknown Service"
}

func (cs ClinicService) Values() []ClinicService {
	return ValidClinicServices
}

func (cs ClinicService) DefaultDuration() int {
	if duration, exists := clinicServiceDurations[cs]; exists {
		return duration
	}
	return 30 // default duration in minutes
}

func normalizeServiceInput(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	input = strings.ReplaceAll(input, " ", "_")
	return input
}

func GetAllAppointmentStatuses() []AppointmentStatus {
	return ValidAppointmentStatuses
}

func GetAllClinicServices() []ClinicService {
	return ValidClinicServices
}

// GetActiveAppointmentStatuses returns statuses that represent active appointments
func GetActiveAppointmentStatuses() []AppointmentStatus {
	return []AppointmentStatus{
		AppointmentStatusPending,
		AppointmentStatusConfirmed,
		AppointmentStatusRescheduled,
	}
}
