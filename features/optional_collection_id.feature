Feature: Collection ID Optional
  Scenario: The one where the collection ID is not sent in the meta data
    Given dp-files-api does not have a file "/data/populations.csv" registered
    And the file meta-data is:
      | isPublishable      | true                                                                      |
      | title              | The number of people                                                      |
      | resumableTotalSize | 14794                                                                     |
      | licence            | OGL v3                                                                    |
      | licenceUrl         | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
    And the data file "populations.csv" with content:
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
