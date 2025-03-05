# Analytics Interceptor for Optimizely Agent

This interceptor captures HTTP requests to and from the Optimizely Agent and sends analytics data to Google Analytics.

## Features

- Tracks API usage patterns
- Captures request and response metrics
- Sends data to Google Analytics (GA4)
- Customizable tracking parameters

## Configuration

Add the following to your `config.yaml` file:

```yaml
server:
  interceptors:
    analytics:
      trackingID: "G-XXXXXXXXXX"  # Your Google Analytics tracking ID
      enabled: true               # Set to false to disable tracking
      endpointURL: ""             # Optional: override the default GA endpoint
```

## Implementation Details

The interceptor captures the following information:
- Request path and method
- Response status code
- Response time
- User agent
- IP address

This data is sent to Google Analytics as an event called "api_request".

## Privacy Considerations

Make sure your use of this interceptor complies with applicable privacy laws and regulations, such as GDPR, CCPA, etc. Consider adding appropriate privacy disclosures to your applications.
