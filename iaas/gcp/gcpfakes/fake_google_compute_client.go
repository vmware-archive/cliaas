// This file was generated by counterfeiter
package gcpfakes

import (
	"sync"
	"time"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
	compute "google.golang.org/api/compute/v1"
)

type FakeGoogleComputeClient struct {
	ListStub        func(project string, zone string) (*compute.InstanceList, error)
	listMutex       sync.RWMutex
	listArgsForCall []struct {
		project string
		zone    string
	}
	listReturns struct {
		result1 *compute.InstanceList
		result2 error
	}
	listReturnsOnCall map[int]struct {
		result1 *compute.InstanceList
		result2 error
	}
	DeleteStub        func(project string, zone string, instanceName string) (*compute.Operation, error)
	deleteMutex       sync.RWMutex
	deleteArgsForCall []struct {
		project      string
		zone         string
		instanceName string
	}
	deleteReturns struct {
		result1 *compute.Operation
		result2 error
	}
	deleteReturnsOnCall map[int]struct {
		result1 *compute.Operation
		result2 error
	}
	InsertStub        func(project string, zone string, instance *compute.Instance) (*compute.Operation, error)
	insertMutex       sync.RWMutex
	insertArgsForCall []struct {
		project  string
		zone     string
		instance *compute.Instance
	}
	insertReturns struct {
		result1 *compute.Operation
		result2 error
	}
	insertReturnsOnCall map[int]struct {
		result1 *compute.Operation
		result2 error
	}
	ImageInsertStub        func(project string, image *compute.Image, timeout time.Duration) (*compute.Operation, error)
	imageInsertMutex       sync.RWMutex
	imageInsertArgsForCall []struct {
		project string
		image   *compute.Image
		timeout time.Duration
	}
	imageInsertReturns struct {
		result1 *compute.Operation
		result2 error
	}
	imageInsertReturnsOnCall map[int]struct {
		result1 *compute.Operation
		result2 error
	}
	StopStub        func(project string, zone string, instanceName string) (*compute.Operation, error)
	stopMutex       sync.RWMutex
	stopArgsForCall []struct {
		project      string
		zone         string
		instanceName string
	}
	stopReturns struct {
		result1 *compute.Operation
		result2 error
	}
	stopReturnsOnCall map[int]struct {
		result1 *compute.Operation
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGoogleComputeClient) List(project string, zone string) (*compute.InstanceList, error) {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
		project string
		zone    string
	}{project, zone})
	fake.recordInvocation("List", []interface{}{project, zone})
	fake.listMutex.Unlock()
	if fake.ListStub != nil {
		return fake.ListStub(project, zone)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.listReturns.result1, fake.listReturns.result2
}

func (fake *FakeGoogleComputeClient) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeGoogleComputeClient) ListArgsForCall(i int) (string, string) {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return fake.listArgsForCall[i].project, fake.listArgsForCall[i].zone
}

func (fake *FakeGoogleComputeClient) ListReturns(result1 *compute.InstanceList, result2 error) {
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 *compute.InstanceList
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) ListReturnsOnCall(i int, result1 *compute.InstanceList, result2 error) {
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 *compute.InstanceList
			result2 error
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 *compute.InstanceList
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) Delete(project string, zone string, instanceName string) (*compute.Operation, error) {
	fake.deleteMutex.Lock()
	ret, specificReturn := fake.deleteReturnsOnCall[len(fake.deleteArgsForCall)]
	fake.deleteArgsForCall = append(fake.deleteArgsForCall, struct {
		project      string
		zone         string
		instanceName string
	}{project, zone, instanceName})
	fake.recordInvocation("Delete", []interface{}{project, zone, instanceName})
	fake.deleteMutex.Unlock()
	if fake.DeleteStub != nil {
		return fake.DeleteStub(project, zone, instanceName)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.deleteReturns.result1, fake.deleteReturns.result2
}

func (fake *FakeGoogleComputeClient) DeleteCallCount() int {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return len(fake.deleteArgsForCall)
}

func (fake *FakeGoogleComputeClient) DeleteArgsForCall(i int) (string, string, string) {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return fake.deleteArgsForCall[i].project, fake.deleteArgsForCall[i].zone, fake.deleteArgsForCall[i].instanceName
}

func (fake *FakeGoogleComputeClient) DeleteReturns(result1 *compute.Operation, result2 error) {
	fake.DeleteStub = nil
	fake.deleteReturns = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) DeleteReturnsOnCall(i int, result1 *compute.Operation, result2 error) {
	fake.DeleteStub = nil
	if fake.deleteReturnsOnCall == nil {
		fake.deleteReturnsOnCall = make(map[int]struct {
			result1 *compute.Operation
			result2 error
		})
	}
	fake.deleteReturnsOnCall[i] = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) Insert(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
	fake.insertMutex.Lock()
	ret, specificReturn := fake.insertReturnsOnCall[len(fake.insertArgsForCall)]
	fake.insertArgsForCall = append(fake.insertArgsForCall, struct {
		project  string
		zone     string
		instance *compute.Instance
	}{project, zone, instance})
	fake.recordInvocation("Insert", []interface{}{project, zone, instance})
	fake.insertMutex.Unlock()
	if fake.InsertStub != nil {
		return fake.InsertStub(project, zone, instance)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.insertReturns.result1, fake.insertReturns.result2
}

func (fake *FakeGoogleComputeClient) InsertCallCount() int {
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	return len(fake.insertArgsForCall)
}

func (fake *FakeGoogleComputeClient) InsertArgsForCall(i int) (string, string, *compute.Instance) {
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	return fake.insertArgsForCall[i].project, fake.insertArgsForCall[i].zone, fake.insertArgsForCall[i].instance
}

func (fake *FakeGoogleComputeClient) InsertReturns(result1 *compute.Operation, result2 error) {
	fake.InsertStub = nil
	fake.insertReturns = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) InsertReturnsOnCall(i int, result1 *compute.Operation, result2 error) {
	fake.InsertStub = nil
	if fake.insertReturnsOnCall == nil {
		fake.insertReturnsOnCall = make(map[int]struct {
			result1 *compute.Operation
			result2 error
		})
	}
	fake.insertReturnsOnCall[i] = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) ImageInsert(project string, image *compute.Image, timeout time.Duration) (*compute.Operation, error) {
	fake.imageInsertMutex.Lock()
	ret, specificReturn := fake.imageInsertReturnsOnCall[len(fake.imageInsertArgsForCall)]
	fake.imageInsertArgsForCall = append(fake.imageInsertArgsForCall, struct {
		project string
		image   *compute.Image
		timeout time.Duration
	}{project, image, timeout})
	fake.recordInvocation("ImageInsert", []interface{}{project, image, timeout})
	fake.imageInsertMutex.Unlock()
	if fake.ImageInsertStub != nil {
		return fake.ImageInsertStub(project, image, timeout)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.imageInsertReturns.result1, fake.imageInsertReturns.result2
}

func (fake *FakeGoogleComputeClient) ImageInsertCallCount() int {
	fake.imageInsertMutex.RLock()
	defer fake.imageInsertMutex.RUnlock()
	return len(fake.imageInsertArgsForCall)
}

func (fake *FakeGoogleComputeClient) ImageInsertArgsForCall(i int) (string, *compute.Image, time.Duration) {
	fake.imageInsertMutex.RLock()
	defer fake.imageInsertMutex.RUnlock()
	return fake.imageInsertArgsForCall[i].project, fake.imageInsertArgsForCall[i].image, fake.imageInsertArgsForCall[i].timeout
}

func (fake *FakeGoogleComputeClient) ImageInsertReturns(result1 *compute.Operation, result2 error) {
	fake.ImageInsertStub = nil
	fake.imageInsertReturns = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) ImageInsertReturnsOnCall(i int, result1 *compute.Operation, result2 error) {
	fake.ImageInsertStub = nil
	if fake.imageInsertReturnsOnCall == nil {
		fake.imageInsertReturnsOnCall = make(map[int]struct {
			result1 *compute.Operation
			result2 error
		})
	}
	fake.imageInsertReturnsOnCall[i] = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) Stop(project string, zone string, instanceName string) (*compute.Operation, error) {
	fake.stopMutex.Lock()
	ret, specificReturn := fake.stopReturnsOnCall[len(fake.stopArgsForCall)]
	fake.stopArgsForCall = append(fake.stopArgsForCall, struct {
		project      string
		zone         string
		instanceName string
	}{project, zone, instanceName})
	fake.recordInvocation("Stop", []interface{}{project, zone, instanceName})
	fake.stopMutex.Unlock()
	if fake.StopStub != nil {
		return fake.StopStub(project, zone, instanceName)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.stopReturns.result1, fake.stopReturns.result2
}

func (fake *FakeGoogleComputeClient) StopCallCount() int {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return len(fake.stopArgsForCall)
}

func (fake *FakeGoogleComputeClient) StopArgsForCall(i int) (string, string, string) {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return fake.stopArgsForCall[i].project, fake.stopArgsForCall[i].zone, fake.stopArgsForCall[i].instanceName
}

func (fake *FakeGoogleComputeClient) StopReturns(result1 *compute.Operation, result2 error) {
	fake.StopStub = nil
	fake.stopReturns = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) StopReturnsOnCall(i int, result1 *compute.Operation, result2 error) {
	fake.StopStub = nil
	if fake.stopReturnsOnCall == nil {
		fake.stopReturnsOnCall = make(map[int]struct {
			result1 *compute.Operation
			result2 error
		})
	}
	fake.stopReturnsOnCall[i] = struct {
		result1 *compute.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleComputeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	fake.imageInsertMutex.RLock()
	defer fake.imageInsertMutex.RUnlock()
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeGoogleComputeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ gcp.GoogleComputeClient = new(FakeGoogleComputeClient)
