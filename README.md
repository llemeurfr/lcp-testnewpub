Test the insertion of multiple records in an LCP License Server.

Usage: command line tool

> lcp-testnewpub

with the following parameters:

- url: LCP License Server URL, e.g. "https://provider.org/license-server:8089"
- tick: Interval between two calls to the LCP License Server in ms, 100 ms by default.
- testtime: Total testing time in ms, 4000 ms by default.

Messages: 

If there is no error, you'll see the total time the process took and the number of records aded to the License Server database. 

"Process took 4.103960042s counter hit 41"

Different error details will be displayed if the test cannot be achieved.


