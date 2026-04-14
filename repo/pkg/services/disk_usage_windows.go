//go:build windows

package services

// getDiskUsagePercent returns 0 on Windows (not the target platform).
func getDiskUsagePercent() (float64, error) {
	return 0, nil
}
