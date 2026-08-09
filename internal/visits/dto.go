package visits

import "time"

type DTO struct {
	ID           int64      `json:"id"`
	VisitorID    int64      `json:"visitor_id"`
	VisitTypeID  *int64     `json:"visit_type_id"`
	DepartmentID *int64     `json:"department_id"`
	HostName     *string    `json:"host_name"`
	Purpose      *string    `json:"purpose"`
	ExpectedTime *time.Time `json:"expected_time"`
	Status       string     `json:"status"`
	BadgeNumber  *string    `json:"badge_number"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
	CheckedOutAt *time.Time `json:"checked_out_at"`
	CancelReason *string    `json:"cancel_reason,omitempty"`
}

func ToDTO(v *Visit) DTO {
	return DTO{
		ID: v.ID, VisitorID: v.VisitorID, VisitTypeID: v.VisitTypeID, DepartmentID: v.DepartmentID,
		HostName: v.HostName, Purpose: v.Purpose, ExpectedTime: v.ExpectedTime, Status: string(v.Status),
		BadgeNumber: v.BadgeNumber, CheckedInAt: v.CheckedInAt, CheckedOutAt: v.CheckedOutAt,
		CancelReason: v.CancelReason,
	}
}

// CreateInput is the explicit allow-list for scheduling/registering a
// visit. CheckInNow lets the front desk do "walk-in arrives -> create +
// check in + badge" as a single call, instead of forcing two round trips
// for the common case.
type CreateInput struct {
	VisitorID    int64      `json:"visitor_id"`
	VisitTypeID  *int64     `json:"visit_type_id"`
	DepartmentID *int64     `json:"department_id"`
	HostName     *string    `json:"host_name"`
	Purpose      *string    `json:"purpose"`
	ExpectedTime *time.Time `json:"expected_time"`
	CheckInNow   bool       `json:"check_in_now"`
}

type CancelInput struct {
	Reason string `json:"reason"`
}

// BadgeDTO is what a security guard's scanner/lookup gets back — enough to
// verify legitimacy at a glance, nothing more (no ID number, no contact
// info leaked into a badge-scan response).
type BadgeDTO struct {
	VisitID      int64      `json:"visit_id"`
	VisitorName  string     `json:"visitor_name"`
	CompanyName  *string    `json:"company_name"`
	DepartmentID *int64     `json:"department_id"`
	HostName     *string    `json:"host_name"`
	Status       string     `json:"status"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
	BadgeNumber  *string    `json:"badge_number"`
}
