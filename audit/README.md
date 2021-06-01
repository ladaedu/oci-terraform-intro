Get audit
=========
In OCI, it is possible to get all events audit, for given compartment.
Here are scripts that I used to check that students are doing their work.


Prerequisities
--------------
- have OCI-CLI installed: `pip3 install oci-cli` and setup
- have `jq` installed


How to run the scripts
----------------------
- edit `get-students-audit.sh`
  - set oci-cli profile
  - set compartments
  - set given date ranges (it is good not to span more than 1 day, as it is faster)
- run the script to retrieve audit logs in JSON format

    ```
    ./get-students-audit.sh
    ```

- if you add some compartment/dates, just run the script again, and it will fetch just the new logs
- filter just non-GET events, and format it from JSON to text:

    ```
    ./jq-audit-raw-filter-no-gets.sh *log|./jq-audit-raw-format.sh
    ```
