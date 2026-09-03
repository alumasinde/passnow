package visits

import "time"

type DTO struct {
	ID int64 `json:"id"`
	VisitorID int64 `json:"visitor_id"`
	EntrySource string `json:"entry_source"`
	VisitorName string `json:"visitor_name,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	VisitTypeID *int64 `json:"visit_type_id"`
	VisitTypeName string `json:"visit_type_name,omitempty"`
	DepartmentID *int64 `json:"department_id"`
	DepartmentName string `json:"department_name,omitempty"`
	HostName *string `json:"host_name"`
	Purpose *string `json:"purpose"`
	ExpectedTime *time.Time `json:"expected_time"`
	ExpectedDepartureAt *time.Time `json:"expected_departure_at"`
	Status string `json:"status"`
	BadgeNumber *string `json:"badge_number"`
	CheckedInAt *time.Time `json:"checked_in_at"`
	CheckedOutAt *time.Time `json:"checked_out_at"`
	CancelReason *string `json:"cancel_reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	CanCheckIn bool `json:"can_check_in"`
	CanCheckOut bool `json:"can_check_out"`
	CanCancel bool `json:"can_cancel"`
	Overstayed bool `json:"overstayed"`
}

func ToDTO(v *Visit) DTO {
	return DTO{
		ID:v.ID,VisitorID:v.VisitorID,EntrySource:string(v.EntrySource),VisitTypeID:v.VisitTypeID,DepartmentID:v.DepartmentID,
		HostName:v.HostName,Purpose:v.Purpose,ExpectedTime:v.ExpectedTime,ExpectedDepartureAt:v.ExpectedDepartureAt,Status:string(v.Status),
		BadgeNumber:v.BadgeNumber,CheckedInAt:v.CheckedInAt,CheckedOutAt:v.CheckedOutAt,
		CancelReason:v.CancelReason,CreatedAt:v.CreatedAt,
		CanCheckIn:v.CanCheckIn(),CanCheckOut:v.CanCheckOut(),CanCancel:v.CanCancel(),
		Overstayed: v.Status==StatusCheckedIn && v.ExpectedTime!=nil && v.ExpectedTime.Before(time.Now().UTC()),
	}
}

type CreateInput struct {
	VisitorID int64 `json:"visitor_id"`
	EntrySource EntrySource `json:"entry_source"`
	VisitTypeID *int64 `json:"visit_type_id"`
	DepartmentID *int64 `json:"department_id"`
	HostName *string `json:"host_name"`
	Purpose *string `json:"purpose"`
	ExpectedTime *time.Time `json:"expected_time"`
	ExpectedDepartureAt *time.Time `json:"expected_departure_at"`
	CheckInNow bool `json:"check_in_now"`
}
type CancelInput struct { Reason string `json:"reason"` }

type BadgeDTO struct {
	VisitID int64 `json:"visit_id"`
	VisitorName string `json:"visitor_name"`
	CompanyName *string `json:"company_name"`
	DepartmentID *int64 `json:"department_id"`
	HostName *string `json:"host_name"`
	Status string `json:"status"`
	CheckedInAt *time.Time `json:"checked_in_at"`
	BadgeNumber *string `json:"badge_number"`
}
