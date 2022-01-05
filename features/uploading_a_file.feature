Feature: Uploading a file

    Background:
        Given dp-files-api does not have a file "/data/populations.csv" registered

    Scenario: File ends up in bucket as result of uploading
        Given the data file "populations.csv" with content:
        """
        mark,1
        jon,2
        russ,3
        """
        When I upload the file "populations.csv" with the following form meta-data:
            | path           | /data/populations.csv                                                     |
            | type           | text/csv                                                                  |
            | isPublishable | true                                                                      |
            | collectionId  | 1234-asdfg-54321-qwerty                                                   |
            | title          | The number of people                                                      |
            | sizeInBytes  | 14794                                                                     |
            | licence        | OGL v3                                                                    |
            | licenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
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
          "path": "/data/populations.csv",
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
          "path": "/data/populations.csv",
          "etag": "94f986e8a3806737648dff3aaa84f57f"
        }
        """
    # Need to work out a way of getting the expected etag from can come from? s3 maybe?





