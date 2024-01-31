Feature: Status for a single file

  Background:
    Given dp-files-api has a file with path "testing" and filename "valid" registered with meta-data:
            """
            {
                "path": "testing/valid",
                "is_publishable": true,
                "title": "",
                "size_in_bytes": 100,
                "type": "text/plain",
                "licence": "na",
                "licence_url": "na",
                "state": "UPLOADED",
                "etag": "49c602a643eab72e8dd68d4e7048883b"
            }
            """

  Scenario: Status returns with valid metadata and status messages for bucket
    Given I GET "/upload-new/files/testing/valid/status"
    Then I should receive the following JSON response with status "200":
            """
                {
                  "metadata": {
                      "path": "testing/valid",
                      "is_publishable": true,
                      "title": "",
                      "size_in_bytes": 100,
                      "type": "text/plain",
                      "licence": "na",
                      "licence_url": "na",
                      "state": "UPLOADED",
                      "etag": "49c602a643eab72e8dd68d4e7048883b"
                  },
                  "file_content": {
                      "valid": true
                  }
              }
            """

  Scenario: Status returns with 404 if path doesnt exist
    Given I GET "/upload-new/files/testing/invalid/status"
    Then the HTTP status code should be "404"
