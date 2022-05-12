# spice

It's an auxiliary project that complement others, but technically doesn't belong anywhere

## List of components
* [Uptime]
    * GET /uptime?url=          to find if url exists
    * POST /uptime              to create uptime record

* [webhook] Uptime stats response
    * POST /hook/uptime

* [contact] Public sector contact
    * GET /q/public-sector-contact/?query=<deptName>
    * any 404 query will be logged for further improvement
