# Graph

This package defines the way that all the data connects and how to load it from
the GraphQL API.

## Structure

* _(not implemented)_ A *User*
  * has attributes:
    * name
    * email
    * password (not plaintext, hash, etc.)


* _(not implemented)_ An *Organization*:
  * has attributes:
    * name
    * orgLimit: max number of sub-orgs
      * < 0 means no limit
      * 0 means none
      * >0 is a limit
  * has many *Organizations*
  * has many *Boards*
  * has many *Users* through *Roles*

* _(not implemented)_ A *Role*
  * has a User
  * has a Name

* A *Board*
  * has attributes:
    * title
    * maxVotes: number of allowed votes per participant
    * timer: default amount of time for a timer
  * has many *Notes*
  * has a *lock state* that determines the actions available
    * Locked: changes are not permitted
    * Unlocked: changes are permitted
  * has *Features*
    * OpenForNotes: allows notes to be created
    * OpenForVotes: allows voting to take place

* A *Note*
  * has a body
  * has votes
  * can be merged with other *Notes*
  * has a *Category*

### Common Metadata

All models, unless otherwise specified, have the following common metadata:

* id
* createdAt
* updatedAt
* deletedAt

All dates and times are stored in UTC.
