package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOrganizationsController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewOrganizationsController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestOrganizationsQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermRead),
		),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Organization
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name: "With one org",
			ctx:  defaultCtx,
			records: []*types.Organization{
				types.FixtureOrganization("default"),
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermCreate),
			)),
			records: []*types.Organization{
				types.FixtureOrganization("org1"),
				types.FixtureOrganization("org2"),
			},
			expectedLen: 0,
			storeErr:    nil,
		},
		{
			name:        "Store Failure",
			ctx:         defaultCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			// Mock store methods
			store.On("GetOrganizations", tc.ctx).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.Query(tc.ctx)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestOrganizationsFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermRead),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.Organization
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No name given",
			ctx:             defaultCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureOrganization("org1"),
			argument:        "org1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "org1",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermCreate),
			)),
			record:          types.FixtureOrganization("org1"),
			argument:        "org1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetOrganizationByName", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.record, nil)

			// Exec Query
			result, err := actions.Find(tc.ctx, tc.argument)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Find() to return a record")
		})
	}
}

func TestOrganizationsCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermCreate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermRead),
		),
	)

	badOrg := types.FixtureOrganization("org1")
	badOrg.Name = "I like turtles"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Organization
		fetchResult     *types.Organization
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureOrganization("org1"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchResult:     types.FixtureOrganization("org1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureOrganization("org1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badOrg,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetOrganizationByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateOrganization", mock.Anything, mock.Anything).
				Return(tc.createErr)

			// Exec Query
			err := actions.Create(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestOrganizationsUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermRead),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Organization
		fetchResult     *types.Organization
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureOrganization("org1"),
			fetchResult: types.FixtureOrganization("org1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchResult:     types.FixtureOrganization("org1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchResult:     types.FixtureOrganization("org1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureOrganization("org1"),
			fetchResult:     types.FixtureOrganization("org1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},

		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        types.FixtureOrganization("bad org"),
			fetchResult:     types.FixtureOrganization("bad org"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetOrganizationByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateOrganization", mock.Anything, mock.Anything).
				Return(tc.updateErr)

			// Exec Query
			err := actions.Update(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestOrganizationsDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermDelete),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermCreate),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.Organization
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			argument:    "org1",
			fetchResult: types.FixtureOrganization("org1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     types.FixtureOrganization("org1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     types.FixtureOrganization("org1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        "org1",
			fetchResult:     types.FixtureOrganization("org1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetOrganizationByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteOrganizationByName", mock.Anything, "org1").
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}
