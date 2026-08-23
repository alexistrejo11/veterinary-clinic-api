package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"clinic-vet-api/database/sqlc"
	cust "clinic-vet-api/internal/core/customers"
	emp "clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"

	"github.com/jackc/pgx/v5/pgtype"

	"clinic-vet-api/internal/shared"
	customErr "clinic-vet-api/internal/shared/errors"
)

type SqlcUserRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSqlcUserRepository(queries *sqlc.Queries, pgMap *mapper.SqlcFieldMapper) users.UserRepository {
	return &SqlcUserRepository{
		queries: queries,
		pgMap:   pgMap,
	}
}

// ============================================================================
// FIND OPERATIONS
// ============================================================================

func (r *SqlcUserRepository) FindByID(ctx context.Context, id shared.UserID) (users.User, error) {
	sqlRow, err := r.queries.FindUserByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("id", id.String())
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("%s with ID %d", ErrMsgFindUser, id.Value()), err)
	}

	return r.ToEntity(sqlRow), nil
}

func (r *SqlcUserRepository) FindByEmail(ctx context.Context, email users.Email) (users.User, error) {
	sqlRow, err := r.queries.FindUserByEmail(ctx, r.pgMap.PgText.FromString(email.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("email", email.Value())
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("%s with email %s", ErrMsgFindUserByEmail, email.Value()), err)
	}

	return r.ToEntity(sqlRow), nil
}

func (r *SqlcUserRepository) FindByPhone(ctx context.Context, phone users.PhoneNumber) (users.User, error) {
	sqlRow, err := r.queries.FindUserByPhoneNumber(ctx, r.pgMap.PgText.FromString(phone.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("phone", phone.String())
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("%s with phone %s", ErrMsgFindUserByPhone, phone.String()), err)
	}

	return r.ToEntity(sqlRow), nil
}

func (r *SqlcUserRepository) FindByCustomerID(ctx context.Context, customerID cust.CustomerID) (users.User, error) {
	userRow, err := r.queries.FindUserByCustomerID(ctx, customerID.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("customer_id", customerID.String())
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("failed to find user by customer ID %d", customerID.Value()), err)
	}

	return r.ToEntity(userRow), nil
}

func (r *SqlcUserRepository) FindByEmployeeID(ctx context.Context, employeeID emp.EmployeeID) (users.User, error) {
	userRow, err := r.queries.FindUserByEmployeeID(ctx, employeeID.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("employee_id", employeeID.String())
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("failed to find user by employee ID %d", employeeID.Value()), err)
	}

	return r.ToEntity(userRow), nil
}

func (r *SqlcUserRepository) FindByOAuthProvider(ctx context.Context, provider string, providerID string) (users.User, error) {
	sqlRow, err := r.queries.FindUserByOAuthProvider(ctx, sqlc.FindUserByOAuthProviderParams{
		OauthProvider:   provider,
		OauthProviderID: r.pgMap.PgText.FromString(providerID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, r.notFoundError("oauth_provider", fmt.Sprintf("%s:%s", provider, providerID))
		}
		return users.User{}, r.dbError(OpSelect, fmt.Sprintf("failed to find user by OAuth provider %s:%s", provider, providerID), err)
	}

	return r.ToEntity(sqlRow), nil
}

func (r *SqlcUserRepository) FindSpecification(ctx context.Context, spec users.UserSpecification) (page.Page[users.User], error) {
	// Build params from specification
	params := r.buildSpecParams(spec)

	// Execute query
	userRows, err := r.queries.FindUsersBySpec(ctx, params)
	if err != nil {
		return page.Page[users.User]{}, r.dbError(OpSelect, "failed to find users by specification", err)
	}

	// Count total
	total, err := r.Count(ctx, spec)
	if err != nil {
		return page.Page[users.User]{}, r.dbError(OpCount, "failed to count users by specification", err)
	}

	usersEntities := r.ToEntities(userRows)
	pagReq := page.PaginationRequest{
		Page: int32(spec.Pagination.Number), PageSize: int32(spec.Pagination.Size),
	}
	return page.NewPage(usersEntities, total, pagReq), nil
}

// ============================================================================
// EXISTS OPERATIONS
// ============================================================================

func (r *SqlcUserRepository) ExistsByID(ctx context.Context, id shared.UserID) (bool, error) {
	exists, err := r.queries.ExistsUserByID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, fmt.Sprintf("%s with ID %d", ErrMsgCheckUserExists, id.Value()), err)
	}
	return exists, nil
}

func (r *SqlcUserRepository) ExistsByEmail(ctx context.Context, email users.Email) (bool, error) {
	exists, err := r.queries.ExistsUserByEmail(ctx, r.pgMap.PgText.FromString(email.Value()))
	if err != nil {
		return false, r.dbError(OpSelect, fmt.Sprintf("%s with email %s", ErrMsgCheckUserExists, email), err)
	}
	return exists, nil
}

func (r *SqlcUserRepository) ExistsByPhone(ctx context.Context, phone users.PhoneNumber) (bool, error) {
	exists, err := r.queries.ExistsUserByPhoneNumber(ctx, r.pgMap.PgText.FromString(phone.String()))
	if err != nil {
		return false, r.dbError(OpSelect, fmt.Sprintf("%s with phone %s", ErrMsgCheckUserExists, phone), err)
	}
	return exists, nil
}

func (r *SqlcUserRepository) ExistsByCustomerID(ctx context.Context, customerID cust.CustomerID) (bool, error) {
	exists, err := r.queries.ExistsUserByCustomerID(ctx, customerID.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, fmt.Sprintf("failed to check user existence by customer ID %d", customerID.Value()), err)
	}
	return exists, nil
}

func (r *SqlcUserRepository) ExistsByEmployeeID(ctx context.Context, employeeID emp.EmployeeID) (bool, error) {
	exists, err := r.queries.ExistsUserByEmployeeID(ctx, employeeID.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, fmt.Sprintf("failed to check user existence by employee ID %d", employeeID.Value()), err)
	}
	return exists, nil
}

// ============================================================================
// SAVE, DELETE & RESTORE OPERATIONS
// ============================================================================

func (r *SqlcUserRepository) Save(ctx context.Context, user *users.User) error {
	if user.ID.IsZero() {
		return r.create(ctx, user)
	}
	return r.update(ctx, user)
}

func (r *SqlcUserRepository) SoftDelete(ctx context.Context, id shared.UserID) error {
	if err := r.queries.SoftDeleteUser(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, fmt.Sprintf("%s with ID %d", ErrMsgSoftDeleteUser, id.Value()), err)
	}
	return nil
}

func (r *SqlcUserRepository) HardDelete(ctx context.Context, id shared.UserID) error {
	if err := r.queries.HardDeleteUser(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, fmt.Sprintf("%s with ID %d", ErrMsgHardDeleteUser, id.Value()), err)
	}
	return nil
}

func (r *SqlcUserRepository) RestoreByID(ctx context.Context, id shared.UserID) error {
	if err := r.queries.RestoreUser(ctx, id.Int32()); err != nil {
		return r.dbError(OpUpdate, fmt.Sprintf("failed to restore user with ID %d", id.Value()), err)
	}
	return nil
}

func (r *SqlcUserRepository) Count(ctx context.Context, spec users.UserSpecification) (int64, error) {
	params := r.buildCountParams(spec)
	count, err := r.queries.CountUsersBySpec(ctx, params)
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count users by specification", err)
	}
	return count, nil
}

func (r *SqlcUserRepository) IsDeletedByID(ctx context.Context, id shared.UserID) (bool, error) {
	deleted, err := r.queries.IsDeletedUserByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, r.dbError(OpSelect, "failed to check if user is deleted", err)
	}
	isDeleted, ok := deleted.(bool)
	if !ok {
		return false, r.dbError(OpSelect, "invalid deleted status", fmt.Errorf("unexpected result type %T", deleted))
	}
	return isDeleted, nil
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

func (r *SqlcUserRepository) create(ctx context.Context, user *users.User) error {
	params := r.toCreateParams(*user)
	userCreated, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return r.dbError(OpInsert, ErrMsgCreateUser, err)
	}

	user.SetID(shared.NewUserID(uint(userCreated.ID)))
	return nil
}

func (r *SqlcUserRepository) update(ctx context.Context, user *users.User) error {
	params := r.toUpdateParams(*user)
	_, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return r.dbError(OpUpdate, fmt.Sprintf("%s with ID %d", ErrMsgUpdateUser, user.ID.Value()), err)
	}
	return nil
}

func (r *SqlcUserRepository) buildSpecParams(spec users.UserSpecification) sqlc.FindUsersBySpecParams {
	var id int32
	if len(spec.IDs) > 0 {
		id = spec.IDs[0].Int32()
	}

	var email, phone, role, status, oauthProvider, searchTerm string
	if len(spec.Emails) > 0 {
		email = spec.Emails[0].String()
	}
	if len(spec.Roles) > 0 {
		role = spec.Roles[0].String()
	}
	if len(spec.Statuses) > 0 {
		status = spec.Statuses[0].String()
	}
	if spec.SearchTerm != nil {
		searchTerm = *spec.SearchTerm
	}

	var emailVerified, twoFAEnabled bool
	// Implementar lógica según spec

	return sqlc.FindUsersBySpecParams{
		Column1:  id,
		Column2:  email,
		Column3:  phone,
		Column4:  role,
		Column5:  status,
		Column6:  oauthProvider,
		Column7:  emailVerified,
		Column8:  twoFAEnabled,
		Column9:  searchTerm,
		Column10: convertTimeToPgTimestamp(spec.CreatedAfter),
		Column11: convertTimeToPgTimestamp(spec.CreatedBefore),
		Column12: convertTimeToPgTimestamp(spec.LastLoginAfter),
		Limit:    int32(spec.Pagination.Size),
		Offset:   int32((spec.Pagination.Number - 1) * spec.Pagination.Size),
	}
}

func (r *SqlcUserRepository) buildCountParams(spec users.UserSpecification) sqlc.CountUsersBySpecParams {
	var id int32
	if len(spec.IDs) > 0 {
		id = spec.IDs[0].Int32()
	}

	var email, phone, role, status, oauthProvider, searchTerm string
	if len(spec.Emails) > 0 {
		email = spec.Emails[0].String()
	}
	if len(spec.Roles) > 0 {
		role = spec.Roles[0].String()
	}
	if len(spec.Statuses) > 0 {
		status = spec.Statuses[0].String()
	}
	if spec.SearchTerm != nil {
		searchTerm = *spec.SearchTerm
	}

	var emailVerified, twoFAEnabled bool
	// Implementar lógica según spec

	return sqlc.CountUsersBySpecParams{
		Column1:  id,
		Column2:  email,
		Column3:  phone,
		Column4:  role,
		Column5:  status,
		Column6:  oauthProvider,
		Column7:  emailVerified,
		Column8:  twoFAEnabled,
		Column9:  searchTerm,
		Column10: convertTimeToPgTimestamp(spec.CreatedAfter),
		Column11: convertTimeToPgTimestamp(spec.CreatedBefore),
		Column12: convertTimeToPgTimestamp(spec.LastLoginAfter),
	}
}

// convertTimeToPgTimestamp converts *time.Time to pgtype.Timestamp
func convertTimeToPgTimestamp(t *time.Time) pgtype.Timestamp {
	if t != nil && !t.IsZero() {
		return pgtype.Timestamp{Time: *t, Valid: true}
	}
	return pgtype.Timestamp{Valid: false}
}

// ============================================================================
// MAPPING METHODS
// ============================================================================

func (r *SqlcUserRepository) ToEntity(sqlRow sqlc.User) users.User {
	id := shared.NewUserID(uint(sqlRow.ID))
	role := users.UserRole(sqlRow.Role)
	phoneNumber := r.pgMap.PgText.ToPhoneNumberPtr(sqlRow.PhoneNumber)
	if phoneNumber == nil {
		phoneNumber = &users.PhoneNumber{}
	}
	userStatus := users.UserStatus(sqlRow.Status)

	// Build OAuth token if available
	var oauthToken *users.OAuthToken
	if sqlRow.OauthAccessToken.Valid {
		oauthToken = &users.OAuthToken{
			AccessToken:  sqlRow.OauthAccessToken.String,
			RefreshToken: r.pgMap.PgText.ToString(sqlRow.OauthRefreshToken),
			ExpiresAt:    r.pgMap.PgTimestamptz.ToTimePtr(sqlRow.OauthTokenExpiry),
		}
	}

	// Build 2FA
	twoFAMethod := shared.TwoFactorMethod(sqlRow.TwoFaMethod)
	if sqlRow.TwoFaMethod == "none" || !sqlRow.TwoFaEnabled.Bool {
		twoFAMethod = shared.TwoFactorMethodUnknown
	}

	twoFA := shared.NewTwoFactorAuth(
		sqlRow.TwoFaEnabled.Bool,
		twoFAMethod,
		r.pgMap.PgText.ToString(sqlRow.TwoFaSecret),
		sqlRow.TwoFaBackupCodes,
		r.pgMap.PgTimestamptz.ToTimePtr(sqlRow.TwoFaEnabledAt),
		r.pgMap.PgTimestamptz.ToTimePtr(sqlRow.Last2faCodeUsedAt),
	)

	user := users.User{
		Email:           users.NewEmailNoErr(r.pgMap.PgText.ToString(sqlRow.Email)),
		Role:            role,
		Status:          userStatus,
		PhoneNumber:     *phoneNumber,
		HashedPassword:  r.pgMap.PgText.ToString(sqlRow.HashedPassword),
		LastLoginAt:     r.pgMap.PgTimestamptz.ToTimePtr(sqlRow.LastLogin),
		TwoFactorAuth:   twoFA,
		OAuthProvider:   sqlRow.OauthProvider,
		OAuthProviderID: r.pgMap.PgText.ToStringPtr(sqlRow.OauthProviderID),
		OAuthToken:      oauthToken,
		EmailVerified:   sqlRow.EmailVerified.Bool,
		LoginAttempts:   int(sqlRow.LoginAttempts.Int32),
		LockedUntil:     r.pgMap.PgTimestamptz.ToTimePtr(sqlRow.LockedUntil),
		Profile: users.Profile{
			Name:        sqlRow.Name,
			Gender:      shared.PersonGender(r.pgMap.PgText.ToString(sqlRow.Gender)),
			PhotoURL:    r.pgMap.PgText.ToString(sqlRow.PhotoUrl),
			Bio:         r.pgMap.PgText.ToString(sqlRow.Bio),
			DateOfBirth: r.pgMap.PgDate.ToTimePtr(sqlRow.DateOfBirth),
		},
	}
	user.SetID(id)
	user.SetTimeStamps(sqlRow.CreatedAt.Time, sqlRow.UpdatedAt.Time)
	return user
}

func (r *SqlcUserRepository) ToEntities(sqlRows []sqlc.User) []users.User {
	users := make([]users.User, len(sqlRows))
	for i, sqlRow := range sqlRows {
		users[i] = r.ToEntity(sqlRow)
	}

	return users
}

func (r *SqlcUserRepository) toCreateParams(user users.User) sqlc.CreateUserParams {
	var oauthAccessToken, oauthRefreshToken pgtype.Text
	var oauthTokenExpiry pgtype.Timestamptz
	var oauthProviderID pgtype.Text

	if user.OAuthToken != nil {
		oauthAccessToken = r.pgMap.PgText.FromString(user.OAuthToken.AccessToken)
		oauthRefreshToken = r.pgMap.PgText.FromString(user.OAuthToken.RefreshToken)
		oauthTokenExpiry = r.pgMap.PgTimestamptz.FromTimePtr(user.OAuthToken.ExpiresAt)
	}

	if user.OAuthProviderID != nil {
		oauthProviderID = r.pgMap.PgText.FromString(*user.OAuthProviderID)
	}

	twoFAMethod := "none"
	if user.TwoFactorAuth.IsEnabled {
		twoFAMethod = string(user.TwoFactorAuth.Method)
	}

	return sqlc.CreateUserParams{
		Email:                       r.pgMap.PgText.FromString(user.Email.String()),
		PhoneNumber:                 r.pgMap.PgText.FromString(user.PhoneNumber.String()),
		HashedPassword:              r.pgMap.PgText.FromString(user.HashedPassword),
		Role:                        user.Role.String(),
		Status:                      user.Status.String(),
		Name:                        user.Profile.Name,
		Gender:                      r.pgMap.PgText.FromString(string(user.Profile.Gender)),
		PhotoUrl:                    r.pgMap.PgText.FromString(user.Profile.PhotoURL),
		Bio:                         r.pgMap.PgText.FromString(user.Profile.Bio),
		DateOfBirth:                 r.pgMap.PgDate.FromTimePtr(user.Profile.DateOfBirth),
		OauthProvider:               user.OAuthProvider,
		OauthProviderID:             oauthProviderID,
		OauthAccessToken:            oauthAccessToken,
		OauthRefreshToken:           oauthRefreshToken,
		OauthTokenExpiry:            oauthTokenExpiry,
		EmailVerified:               pgtype.Bool{Bool: user.EmailVerified, Valid: true},
		TwoFaMethod:                 twoFAMethod,
		TwoFaSecret:                 r.pgMap.PgText.FromString(user.TwoFactorAuth.Secret),
		TwoFaEnabled:                pgtype.Bool{Bool: user.TwoFactorAuth.IsEnabled, Valid: true},
		TwoFaEnabledAt:              r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.EnabledAt),
		TwoFaBackupCodes:            user.TwoFactorAuth.BackupCodes,
		TwoFaBackupCodesGeneratedAt: r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.EnabledAt),
		Last2faCodeUsedAt:           r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.LastUsedAt),
		LastLogin:                   r.pgMap.PgTimestamptz.FromTimePtr(user.LastLoginAt),
		LoginAttempts:               pgtype.Int4{Int32: int32(user.LoginAttempts), Valid: true},
		LockedUntil:                 r.pgMap.PgTimestamptz.FromTimePtr(user.LockedUntil),
	}
}

func (r *SqlcUserRepository) toUpdateParams(user users.User) sqlc.UpdateUserParams {
	var oauthAccessToken, oauthRefreshToken pgtype.Text
	var oauthTokenExpiry pgtype.Timestamptz
	var oauthProviderID pgtype.Text

	if user.OAuthToken != nil {
		oauthAccessToken = r.pgMap.PgText.FromString(user.OAuthToken.AccessToken)
		oauthRefreshToken = r.pgMap.PgText.FromString(user.OAuthToken.RefreshToken)
		oauthTokenExpiry = r.pgMap.PgTimestamptz.FromTimePtr(user.OAuthToken.ExpiresAt)
	}

	if user.OAuthProviderID != nil {
		oauthProviderID = r.pgMap.PgText.FromString(*user.OAuthProviderID)
	}

	twoFAMethod := "none"
	if user.TwoFactorAuth.IsEnabled {
		twoFAMethod = string(user.TwoFactorAuth.Method)
	}

	return sqlc.UpdateUserParams{
		ID:                          int32(user.ID.Value()),
		Email:                       r.pgMap.PgText.FromString(user.Email.String()),
		PhoneNumber:                 r.pgMap.PgText.FromString(user.PhoneNumber.String()),
		HashedPassword:              r.pgMap.PgText.FromString(user.HashedPassword),
		Role:                        user.Role.String(),
		Status:                      user.Status.String(),
		Name:                        user.Profile.Name,
		Gender:                      r.pgMap.PgText.FromString(string(user.Profile.Gender)),
		PhotoUrl:                    r.pgMap.PgText.FromString(user.Profile.PhotoURL),
		Bio:                         r.pgMap.PgText.FromString(user.Profile.Bio),
		DateOfBirth:                 r.pgMap.PgDate.FromTimePtr(user.Profile.DateOfBirth),
		OauthProvider:               user.OAuthProvider,
		OauthProviderID:             oauthProviderID,
		OauthAccessToken:            oauthAccessToken,
		OauthRefreshToken:           oauthRefreshToken,
		OauthTokenExpiry:            oauthTokenExpiry,
		EmailVerified:               pgtype.Bool{Bool: user.EmailVerified, Valid: true},
		TwoFaMethod:                 twoFAMethod,
		TwoFaSecret:                 r.pgMap.PgText.FromString(user.TwoFactorAuth.Secret),
		TwoFaEnabled:                pgtype.Bool{Bool: user.TwoFactorAuth.IsEnabled, Valid: true},
		TwoFaEnabledAt:              r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.EnabledAt),
		TwoFaBackupCodes:            user.TwoFactorAuth.BackupCodes,
		TwoFaBackupCodesGeneratedAt: r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.EnabledAt),
		Last2faCodeUsedAt:           r.pgMap.PgTimestamptz.FromTimePtr(user.TwoFactorAuth.LastUsedAt),
		LastLogin:                   r.pgMap.PgTimestamptz.FromTimePtr(user.LastLoginAt),
		LoginAttempts:               pgtype.Int4{Int32: int32(user.LoginAttempts), Valid: true},
		LockedUntil:                 r.pgMap.PgTimestamptz.FromTimePtr(user.LockedUntil),
	}
}

// ============================================================================
// ERROR HANDLING METHODS
// ============================================================================

func (r *SqlcUserRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableUsers, DriverSQL, fmt.Errorf("%s: %v", message, err))
}

func (r *SqlcUserRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableUsers, DriverSQL)
}

func (r *SqlcUserRepository) wrapConversionError(err error) error {
	return customErr.WrapError(context.Background(), err, OpSelect, TableUsers, DriverSQL, ErrMsgConvertUserToDomain)
}
