// Automatically generated by MockGen. DO NOT EDIT!
// Source: hinter.go

package topology

import (
	gomock "github.com/golang/mock/gomock"
	v1 "kubevirt.io/api/core/v1"
)

// Mock of Hinter interface
type MockHinter struct {
	ctrl     *gomock.Controller
	recorder *_MockHinterRecorder
}

// Recorder for MockHinter (not exported)
type _MockHinterRecorder struct {
	mock *MockHinter
}

func NewMockHinter(ctrl *gomock.Controller) *MockHinter {
	mock := &MockHinter{ctrl: ctrl}
	mock.recorder = &_MockHinterRecorder{mock}
	return mock
}

func (_m *MockHinter) EXPECT() *_MockHinterRecorder {
	return _m.recorder
}

func (_m *MockHinter) TopologyHintsForVMI(vmi *v1.VirtualMachineInstance) (*v1.TopologyHints, TscFrequencyRequirementType, error) {
	ret := _m.ctrl.Call(_m, "TopologyHintsForVMI", vmi)
	ret0, _ := ret[0].(*v1.TopologyHints)
	ret1, _ := ret[1].(TscFrequencyRequirementType)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockHinterRecorder) TopologyHintsForVMI(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TopologyHintsForVMI", arg0)
}

func (_m *MockHinter) IsTscFrequencyRequiredForBoot(vmi *v1.VirtualMachineInstance) bool {
	ret := _m.ctrl.Call(_m, "IsTscFrequencyRequiredForBoot", vmi)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockHinterRecorder) IsTscFrequencyRequiredForBoot(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsTscFrequencyRequiredForBoot", arg0)
}

func (_m *MockHinter) TSCFrequenciesInUse() []int64 {
	ret := _m.ctrl.Call(_m, "TSCFrequenciesInUse")
	ret0, _ := ret[0].([]int64)
	return ret0
}

func (_mr *_MockHinterRecorder) TSCFrequenciesInUse() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TSCFrequenciesInUse")
}

func (_m *MockHinter) LowestTSCFrequencyOnCluster() (int64, error) {
	ret := _m.ctrl.Call(_m, "LowestTSCFrequencyOnCluster")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockHinterRecorder) LowestTSCFrequencyOnCluster() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LowestTSCFrequencyOnCluster")
}
