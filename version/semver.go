package version

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer is a struct to store a version and do some functions on it
type SemVer struct {
	Major   int
	Minor   int
	Patch   int
	Labels  []string
	numbers []int
}

// NewSemVer creates a SemVer from a version in form of string
func NewSemVer(version string) (*SemVer, error) {
	if len(version) == 0 {
		return nil, fmt.Errorf("version cannot be an empty string")
	}

	semver := &SemVer{}

	l := strings.Split(version, "-")
	if len(l) == 1 {
		semver.Labels = []string{}
	} else {
		semver.Labels = l[1:]
	}

	version = l[0]

	numbersStr := strings.Split(version, ".")
	numbers := make([]int, len(numbersStr))

	errs := []error{}
	for i, nStr := range numbersStr {
		if nStr == "" {
			errs = append(errs, fmt.Errorf("this version has a missing number at position %d", i))
			continue
		}
		n, err := strconv.Atoi(nStr)
		if err != nil {
			errs = append(errs, fmt.Errorf("this version has a number that is not a number at position %d (%s)", i, nStr))
			continue
		}
		numbers[i] = n
	}

	var errStr string
	if len(errs) > 0 {
		errStr = "Errors: " + errs[0].Error()
		for _, e := range errs[1:] {
			errStr = errStr + ", " + e.Error()
		}
		return nil, fmt.Errorf(errStr)
	}

	if len(numbers) > 0 {
		semver.Major = numbers[0]
	}
	if len(numbers) > 1 {
		semver.Minor = numbers[1]
	}
	if len(numbers) > 2 {
		semver.Patch = numbers[2]
	}
	semver.numbers = numbers

	return semver, nil
}

func (v *SemVer) String() string {
	var number string
	if len(v.numbers) > 0 {
		number = fmt.Sprintf("%d", v.numbers[0])
		for _, n := range v.numbers[1:] {
			number = fmt.Sprintf("%s.%d", number, n)
		}
	}

	var labels string
	if len(v.Labels) > 0 {
		labels = "-" + strings.Join(v.Labels, "-")
	}

	return fmt.Sprintf("%s%s", number, labels)
}

// Diff returns the difference between two versions
func (v *SemVer) Diff(ver *SemVer) *SemVer {
	major := v.Major - ver.Major
	minor := v.Minor - ver.Minor
	patch := v.Patch - ver.Patch
	numbers := []int{major, minor, patch}

	labelsV := strings.Join(v.Labels, "-")
	labelsVer := strings.Join(ver.Labels, "-")
	var labels []string
	if labelsV != labelsVer {
		labels = []string{fmt.Sprintf("differ (%s != %s)", labelsV, labelsVer)}
	}

	return &SemVer{
		Major:   major,
		Minor:   minor,
		Patch:   patch,
		numbers: numbers,
		Labels:  labels,
	}
}

// Compare compares two versions in form of SemVer. Rules are:
// v > ver  :  1
// v < ver  : -1
// v == ver :  0
func (v *SemVer) Compare(ver *SemVer) int {
	diff := v.Diff(ver)
	for _, n := range diff.numbers {
		if n < 0 {
			return -1
		}
		if n > 0 {
			return 1
		}
	}
	return 0
}

// Compare compares two version in string form
func Compare(version1, version2 string) (int, error) {
	v1, err := NewSemVer(version1)
	if err != nil {
		return 0, err
	}
	v2, err := NewSemVer(version2)
	if err != nil {
		return 0, err
	}

	return v1.Compare(v2), nil
}

// EQ returns true if the given version is equal to this version
func (v *SemVer) EQ(ver *SemVer) bool {
	return v.String() == ver.String()
}

// NE returns true if the given version is equal to this version
func (v *SemVer) NE(ver *SemVer) bool {
	return !v.EQ(ver)
}

// GT returns true if the given version is greater than this version
func (v *SemVer) GT(ver *SemVer) bool {
	d := v.Compare(ver)
	return d == 1
}

// GE returns true if the given version is greater than or equal to this version
func (v *SemVer) GE(ver *SemVer) bool {
	d := v.Compare(ver)
	return d == 1 || d == 0
}

// LT returns true if the given version is less than this version
func (v *SemVer) LT(ver *SemVer) bool {
	d := v.Compare(ver)
	return d == -1
}

// LE returns true if the given version is less than or equal to this version
func (v *SemVer) LE(ver *SemVer) bool {
	d := v.Compare(ver)
	return d == -1 || d == 0
}

// In returns true if the given version is in the given range
func (v *SemVer) In(min, max *SemVer) bool {
	return v.GE(min) && v.LE(max)
}
