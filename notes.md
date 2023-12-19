## Packages

### hamdb
This is a client to hamdb.org, a database of amateur radio operators. It is
used to look up the callsign of a station. It is used by the `callsign`
command.

It should have the following features:

* [ ] Query hamdb.org for a callsign
* [ ] Cache results for a configurable amount of time
* [ ] Cache results in a sqlite database in configurable location
* [ ] Use the sqlite database to look up callsigns

Example response:

```
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "KQ4JXI",
      "class": "T",
      "expires": "07/15/2033",
      "status": "A",
      "grid": "EL96vg",
      "lat": "26.2711019",
      "lon": "-80.2457015",
      "fname": "Ryan",
      "mi": "G",
      "name": "Faerman",
      "suffix": "",
      "addr1": "3211 NW 89th Way",
      "addr2": "Coral Springs",
      "state": "FL",
      "zip": "33065",
      "country": "United States"
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

### TUI
The ui should be a TUI. It should have the following features:

* [ ] Display the current time and date
* [ ] Display the callsign of the operators
* [ ] Display the net frequency details, repeater, offset, etc.
* [ ] Display the preamble
* [ ] Display the postamble
* [ ] Display the checked in stations
* [ ] Track if a station has checked in and if they've been acknowledged
* [ ] Display the check in prompt
* [ ] Have a check-in prompt that will auto-complete callsigns and details from hamdb
* [ ] Generate the net report
* [ ] Display the net report
* [ ] Look up previous net reports 
* [ ] Look up previous net sessions
