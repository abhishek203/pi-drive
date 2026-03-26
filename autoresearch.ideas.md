# Autoresearch Ideas

## Completed Optimizations ✓
- [x] Fixed static analysis issue in indexer.go
- [x] Removed unused sftpd package (-316 LOC)
- [x] Removed unused mount helper functions (-113 LOC)  
- [x] Added `-ldflags="-s -w"` for 31% smaller binaries
- [x] Added unit tests for auth, share, config, search, api packages
- [x] Added godoc comments to helpers.go and middleware.go
- [x] Removed unreachable WebDAV, auth, billing, and mount code paths
- [x] Replaced Cobra CLI framework with stdlib flag parsing for a smaller binary

## Future Optimization Opportunities

### Performance Improvements
- Add database connection pooling configuration
- Consider prepared statements for frequently executed queries
- Add caching for frequently accessed data (plans, agent info)
- Cache parsed templates in templates package

### Code Quality
- Add more unit tests for:
  - Trash service
  - Activity service
  - Billing service
- Add integration tests for API handlers
- Add benchmarks for critical paths
- Increase test coverage to 50%+

### Security Improvements
- Rate limit implementation could use Redis for distributed rate limiting
- Add request logging middleware
- Add request ID tracing

### Documentation
- Add godoc comments to remaining exported functions
- Add API documentation (OpenAPI/Swagger)
- Add CONTRIBUTING.md
