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
    And encryption key will be "0aaf0aaf0aaf0aaf0aaf0aaf0aaf0aaf"

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
    And the path "/data/populations.csv" should be available in the S3 bucket matching content using encryption key "0aaf0aaf0aaf0aaf0aaf0aaf0aaf0aaf":
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
          "etag": "014e5ce3eb6d33344da544b0831140b4"
        }
        """
    And the encryption key "0aaf0aaf0aaf0aaf0aaf0aaf0aaf0aaf" should be stored against file "data/populations.csv"

  Scenario: File upload is marked as started when first chunk is uploaded
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | countries.csv      |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
      | path                 | data               |
    Then the HTTP status code should be "200"
    And the file upload should be marked as started using payload:
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
          "etag": "2da887c94287adc1d46fdc336f8d7852"
        }
        """
    And the stored file "data/countries.csv" should match the sent file "features/countries.csv" using encryption key "0aaf0aaf0aaf0aaf0aaf0aaf0aaf0aaf"
    But the file upload should not have been registered again

