Installatie van `qtechng` op de ontwikkelserver
###################################################



Inleiding
=================

De *baseline* om `qtechng` te installeren op een ontwikkelserver is een server met de volgende eigenschappen:

    - Een passende versie van het Red Hat operating system.
    - Uitgerust met SSH 
    - een up-to-date versie van git

De volgende registry waarden dienen te worden gezet:

-------------------------------    ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
Key                                Value
-------------------------------    ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
qtechng-type                       "B"
qtechng-exe                        "qtechng"
qtechng-max-parallel               "32"
qtechng-repository-dir             "/library/repository"
qtechng-test                       "test-entry"
qtechng-user                       "usystem"
qtechng-unique-ext                 ".m .x"
qtechng-version                    "0.00"
qtechng-copy-exe                   "[\"rsync\", \"-ai\", \"--delete\", \"--exclude=source/.hg\",  \"--exclude=source/.git\", \"/library/repository/{versionsource}/\", \"/library/repository/{versiontarget}\"]"
m-import-auto-exe                  [\"qtechng\", \"fs\", \"store\", \"/library/tmp/tomumps\", \"--append\"]"
lock-dir                           "/run/lock/subsys"
-------------------------------    ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------


De volgende directories dienen te worden uitgerust met een `setgid` bit:

    - /library/tmp
    - /library/repository

De gebruiker `usystem` dient te bestaan en tot de group `db` te behoren.

De `qtechng` binary
=======================

De `qtechng-binary` heeft als basename `qtechng`.

De directory waarin dit bestand staat, wordt aangegeven door de registry waarde `qtechng-binary-dir`.

Deze file moet worden geplaatst in de directory, gegeven door de registry waarde `bindir`.

De *owner* van dit bestand moet `root` zijn, de *group* is `db`, het `setuid` bit moet worden gezet.