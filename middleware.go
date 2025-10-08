// Package diffy provides middleware functionality for validation operations
package diffy

type ValidationMiddleware func([]ValidationFinding, error) ([]ValidationFinding, error)

func LoggingMiddleware(logger Logger) ValidationMiddleware {
	return func(findings []ValidationFinding, err error) ([]ValidationFinding, error) {
		if err != nil {
			logger.Logf("Validation failed: %v", err)
			return findings, err
		}

		if len(findings) == 0 {
			logger.Logf("Validation completed successfully - no issues found")
		} else {
			logger.Logf("Validation completed - found %d issues", len(findings))
		}

		return findings, err
	}
}

func ApplyMiddleware(findings []ValidationFinding, err error, middleware ValidationMiddleware) ([]ValidationFinding, error) {
	return middleware(findings, err)
}
