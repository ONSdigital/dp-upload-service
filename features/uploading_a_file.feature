Feature: Uploading a file

  Background:
    Given dp-files-api does not have a file "/data/populations.csv" registered
    And the file meta-data is:
      | isPublishable      | true                                                                      |
      | collectionId       | 1234-asdfg-54321-qwerty                                                   |
      | title              | The number of people                                                      |
      | resumableTotalSize | 14794                                                                     |
      | licence            | OGL v3                                                                    |
      | licenceUrl         | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |

  Scenario: File ends up in bucket as result of uploading in a single chunk
    Given the data file "populations.csv" with content:
        """
        mark,1
        jon,2
        russ,3
        """
    When I upload the file "test-data/populations.csv" with the following form resumable parameters:
      | resumableFilename    | populations.csv      |
      | resumableType        | text/csv             |
      | resumableTotalChunks | 1                    |
      | resumableChunkNumber | 1                    |
      | path                 | data                 |
    Then the HTTP status code should be "201"
    And the path "/data/populations.csv" should be available in the S3 bucket:
        """
        mark,1
        jon,2
        russ,3
        """
    And the file upload should be marked as started using payload:
        """
        {
          "path": "data/populations.csv",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The number of people",
          "size_in_bytes": 14794,
          "type": "text/csv",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    And the file "data/populations.csv" should be marked as uploaded using payload:
        """
        {
          "state": "UPLOADED",
          "etag": "32204b5b34cd2e635d6d69a443916584-1"
        }
        """

  Scenario: File upload is marked as started when first chunk is uploaded
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | countries.csv      |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
      | path                 | data               |
    Then the HTTP status code should be "200"
    And the file upload should be marked as created using payload:
        """
        {
        "path": "data/countries.csv",
        "is_publishable": true,
        "collection_id": "1234-asdfg-54321-qwerty",
        "title": "The number of people",
        "size_in_bytes": 14794,
        "type": "text/csv",
        "licence": "OGL v3",
        "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    But the file should not be marked as uploaded

  Scenario: File ends up in bucket as result of uploading second chunk
    Given the 1st part of the file "features/countries.csv" has been uploaded with resumable parameters:
      | resumableFilename    | countries.csv      |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
      | path                 | data               |
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | countries.csv      |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 2                  |
      | path                 | data               |
    Then the HTTP status code should be "201"
    And the file "data/countries.csv" should be marked as uploaded using payload:
        """
        {
          "state": "UPLOADED",
          "etag": "5d363622e69e68f5baadc6cef11cbd9f-2"
        }
        """
    And the stored file "data/countries.csv" should match the sent file "features/countries.csv"
    But the file upload should not have been registered again

    Scenario: The one where a single chunk file is uploaded using an authorisation header
      Given the data file "authorized.csv" with content:
        """
        brian,1
        russ,2
        """
      When I upload the file "test-data/authorized.csv" with the following form resumable parameters:
        | resumableFilename    | authorized.csv       |
        | resumableType        | text/csv             |
        | resumableTotalChunks | 1                    |
        | resumableChunkNumber | 1                    |
        | path                 | data                 |
      Then the files api POST request should contain a default authorization header
      And the files api PATCH request with path ("data/authorized.csv") should contain a default authorization header
      And the HTTP status code should be "201"

    Scenario: Uploading a file that already exists returns 409 Conflict
      Given dp-files-api has a file with path "data" and filename "populations.csv" registered with meta-data:
        """
        {"path":"data/populations.csv","title":"existing file"}
        """
      And the data file "populations.csv" with content:
        """
        mark,1
        jon,2
        russ,3
        """
      And the file meta-data is:
        | isPublishable      | true                                                                      |
        | collectionId       | 1234-asdfg-54321-qwerty                                                   |
        | title              | The number of people                                                      |
        | resumableTotalSize | 14794                                                                     |
        | licence            | OGL v3                                                                    |
        | licenceUrl         | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      When I upload the file "test-data/populations.csv" with the following form resumable parameters:
        | resumableFilename    | populations.csv |
        | resumableType        | text/csv        |
        | resumableTotalChunks | 1               |
        | resumableChunkNumber | 1               |
        | path                 | data            |
      Then the HTTP status code should be "409"
      And I should receive the following JSON response:
        """
        {"errors":[{"code":"DuplicateFile","description":"resource conflict"}]}
        """


