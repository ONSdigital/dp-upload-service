Feature: Health check

  Scenario:
    When I GET "/health"
    Then the HTTP status code should be "200"

