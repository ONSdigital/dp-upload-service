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
    And encryption key will be "abcdef123456789z"

  Scenario: File ends up in bucket as result of uploading in a single chunk
    Given the data file "populations.csv" with content:
        """
        mark,1
        jon,2
        russ,3
        """
    When I upload the file "test-data/populations.csv" with the following form resumable parameters:
      | resumableFilename    | data/populations.csv |
      | resumableType        | text/csv             |
      | resumableTotalChunks | 1                    |
      | resumableChunkNumber | 1                    |
    Then the HTTP status code should be "200"
    And the path "/data/populations.csv" should be available in the S3 bucket matching content using encryption key "abcdef123456789z":
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
    And the file should be marked as uploaded using payload:
        """
        {
          "path": "data/populations.csv",
          "etag": "5efb5b786de7942b02fb4bfd63c5715d"
        }
        """
    And the encryption key "abcdef123456789z" should be stored against file "data/populations.csv"

  Scenario: File upload is marked as started when first chunk is uploaded
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | data/countries.csv |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
    Then the HTTP status code should be "100"
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

  Scenario: File upload is marked as started when first chunk is uploaded
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | data/countries.csv |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
    Then the HTTP status code should be "100"
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
      | resumableFilename    | data/countries.csv |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 1                  |
    When I upload the file "features/countries.csv" with the following form resumable parameters:
      | resumableFilename    | data/countries.csv |
      | resumableType        | text/csv           |
      | resumableTotalChunks | 2                  |
      | resumableChunkNumber | 2                  |
    Then the HTTP status code should be "200"
    And the file should be marked as uploaded using payload:
        """
        {
          "path": "data/countries.csv",
          "etag": "714df73fd9a27da75dc6c2d16765e868"
        }
        """
    And the stored file "data/countries.csv" should match the sent file "features/countries.csv" using encryption key "abcdef123456789z"
    But the file upload should not have been registered again

