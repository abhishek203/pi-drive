# Autoresearch Ideas

## Future Optimization Opportunities

### Dead Code Removal (Conservative)
- Review if `InvalidateAgentHandler` in webdav_handler.go is needed or can be called
- Review if billing service dead methods are intended for future Stripe integration
- Review if mount service methods are intended for future JuiceFS direct mount feature

### Performance Improvements
- Add database connection pooling configuration
- Consider prepared statements for frequently executed queries
- Add caching for frequently accessed data (plans, agent info)

### Code Quality
- Add more unit tests for:
  - Search service
  - Trash service  
  - Activity service
- Add integration tests for API handlers
- Add benchmarks for critical paths

### Binary Size Reduction
- Review if all dependencies are necessary
- Consider using `-ldflags="-s -w"` for production builds
- Try TinyGo for CLI binary (if compatible)

### Security Improvements
- Rate limit implementation could use Redis for distributed rate limiting
- Add request logging middleware
- Add request ID tracing

### Documentation
- Add godoc comments to exported functions
- Add API documentation (OpenAPI/Swagger)
