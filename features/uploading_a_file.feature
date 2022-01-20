Feature: Uploading a file

    Background:
        Given dp-files-api does not have a file "/data/populations.csv" registered
        And the file meta-data is:
            | isPublishable | true                                                                      |
            | collectionId  | 1234-asdfg-54321-qwerty                                                   |
            | title         | The number of people                                                      |
            | sizeInBytes   | 14794                                                                     |
            | licence       | OGL v3                                                                    |
            | licenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |

    Scenario: File ends up in bucket as result of uploading in a single chunk
        Given the data file "populations.csv" with content:
        """
        mark,1
        jon,2
        russ,3
        """
        When I upload the file "test-data/populations.csv" with the following form resumable parameters:
            | path         | data/populations.csv |
            | type         | text/csv             |
            | totalChunks  | 1                    |
            | currentChunk | 1                    |
        Then the HTTP status code should be "200"
        And the path "/data/populations.csv" should be available in the S3 bucket matching content:
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
          "etag": "94f986e8a3806737648dff3aaa84f57f"
        }
        """

    Scenario: File upload is marked as started when first chunk is uploaded
        When I upload the file "features/countries.csv" with the following form resumable parameters:
            | path         | data/countries.csv |
            | type         | text/csv           |
            | totalChunks  | 2                  |
            | currentChunk | 1                  |
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
            | path         | data/countries.csv |
            | type         | text/csv           |
            | totalChunks  | 2                  |
            | currentChunk | 1                  |
        When I upload the file "features/countries.csv" with the following form resumable parameters:
            | path         | data/countries.csv |
            | type         | text/csv           |
            | totalChunks  | 2                  |
            | currentChunk | 2                  |
        Then the HTTP status code should be "200"
        And the file should be marked as uploaded using payload:
        """
        {
          "path": "data/countries.csv",
          "etag": "d73ab646c2c25b580bb0f7f7dbd3d454"
        }
        """
        And the stored file "data/countries.csv" should match the sent file "features/countries.csv"
        But the file upload should not have been registered again






