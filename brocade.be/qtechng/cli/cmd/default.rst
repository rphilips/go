Default waarden
====================================


`cwd`
------------

De `current working directory` wordt steeds berekend volgens een waterval systeem:

    - specificatie door middel van de :option:`--cwd` vlag
    - actuele `current working directory`

`version`
--------------

Deze waarde wordt steeds berekend volgens een waterval systeem:

   - specificatie door middel van de :option:`--version` vlag
   - geregisteerde bestanden in de `current working directory`: is er juist 1 gebruikte versie dan wordt ook deze genomen
   - de registry waarde `qtechng-version`



