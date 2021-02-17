:command:`qtechng fs`
====================================

Het `fs` commando heeft op zich weinig met Brocade ontwikkeling te maken. 
Deze familie van instructies bevat een aantal systeem onafhankelijke instrumenten om bestanden, 
die zich in het lokale filesysteem bevinden, te manipuleren.

.. warning:: 

    Sommige van deze instructies zijn erg destructief en moeten met 
    de nodige omzichtigheid worden uitgevoerd.

Alle commando's werken op:

    - `W`: Werkstations
    - `B`: Ontwikkelmachine (dev.anet.be)
    - `P`: Productie servers




:command:`qtechng fs awk`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs awk` voert een AWK instructie uit op de bestanden en verwerkt deze `in-place`.

Het eerste argument is een *AWK commando*.

De te behandelen bestanden worden gespecificeerd door:

    - de andere argumenten (een directory argument betekent steeds alle bestanden in de directory)
    - de :option:`recurse` vlag
    - de :option:`pattern` vlag(gen)



Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--delete                     Eens gekopieerd, worden de *sources* ook geschrapt. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs awk '{print $1}' *.txt`









:command:`qtechng fs copy`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs copy` kopieert lokale bestanden.
Het is ook mogelijk om de originele bestanden te verwijderen waardoor het commando
effectief ook een `rename` functie heeft.

Het eerste argument is de *zoek-string*
Het tweede argument is de *vervang-string*

De *zoek-string* wordt gezocht in de *sources* (= de te kopieren bestanden). 
Indien aanwezig, wordt deze vervangen door de *vervang-string*.

De vlag :option:`--regexp` kan er voor zorgen dat de *zoek-string* als een reguliere 
uitdrukking wordt geinterpreteerd. De laat meteen ook toe om de `gevangen` waarden
`$0`, `$1`, ... te gaan gebruiken in de *vervang-string*.



Vervolgens wordt de *source* gekopieerd naar de *target*

Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--check                      Indien gespecificeerd, dan wordt de eerste keer dat de actie wordt uitgevoerd, gevraagd
                             of dit goed werkt. 

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--regexp                     De *zoek-string* wordt als een reguliere uitdrukking behandeld. 

--delete                     Eens gekopieerd, worden de *sources* ook geschrapt. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs copy d macro isad.d thing.d`

:samp:`qtechng fs copy pdf PDF . --recurse`

:samp:`qtechng fs copy '.*' '${0}.bak' . --regexp --recurse --pattern='*.txt'`




:command:`qtechng fs delete`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs delete` schrapt lokale bestanden.

De te behandelen bestanden worden gespecificeerd door:

    - de argumenten (een directory argument betekent steeds alle bestanden in de directory)
    - de :option:`recurse` vlag
    - de :option:`pattern` vlag(gen)

Indien de argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--check                      Indien gespecificeerd, dan wordt de eerste keer dat de actie wordt uitgevoerd, gevraagd
                             of dit goed werkt. 

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*

--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs delete *.bak`

:samp:`qtechng fs delete . --pattern="*.bak" --recurse`





:command:`qtechng fs grep`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs grep` zoekt naar strings, lijn per lijn, in de bestanden

Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--regexp                     De *zoek-string* wordt als een reguliere uitdrukking behandeld. 

--tolower                    De hoofdletters worden omgezet naar kleine letters. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs grep m4_CO . --recurse --pattern='*.m'`


:command:`qtechng fs replace`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs replace` verandert string door andere strings in bestanden en verwerkt deze `in-place`.

Het eerste argument is de *zoek-string*
Het tweede argument is de *vervang-string*

De *zoek-string* wordt gezocht in de *sources* 
Indien aanwezig, wordt deze vervangen door de *vervang-string*.

De vlag :option:`--regexp` kan er voor zorgen dat de *zoek-string* als een reguliere 
uitdrukking wordt geinterpreteerd. De laat meteen ook toe om de `gevangen` waarden
`$0`, `$1`, ... te gaan gebruiken in de *vervang-string*.




Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--regexp                     De *zoek-string* wordt als een reguliere uitdrukking behandeld. 

--delete                     Eens gekopieerd, worden de *sources* ook geschrapt. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs replace brocade Brocade *.txt`



:command:`qtechng fs rstrip`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs rstrip` verwijdert de wUnicode whitespace karakters die op het einde van een lijn staan.


Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs rstrip *.txt`




:command:`qtechng fs sed`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs sed` voert een sed instructie uit op de bestanden en verwerkt deze `in-place`.

Het eerste argument is een *sed commando*.


Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--delete                     Eens gekopieerd, worden de *sources* ook geschrapt. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs sed 's/e/E/g' *.txt`







:command:`qtechng fs setproperty`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs setproperty` past eigendomsrechten en toegangsrechten aan.

`qtechng` moet hiervoor in gepriviligeerde mode werken! (Op Linux betekent dit opstarten met `sudo`)

Het eerste argument is de *pathmode*: `naked`, `process`, `qtech`, `script`, `temp`, `web`, `webdav`

De andere argumenten zijn de bestanden en directories


Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*


--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs setproperty process *.txt`









:command:`qtechng fs store`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs store` kopieert de standard input naar een bestand.

Het eerste argument is de naam van het bestand.


Dit is een *convenience* faciliteit: het verhinddert het gebruik van shell-afhankelijke operatoren (zoals `>` en `>>`)


Opties
~~~~~~~~~~~

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`


--append                     Voeg toe aan het bestand (in plaats van het te overschrijven)


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs store input.txt`



:command:`qtechng fs touch`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs touch` past de *last modification date* aan van de gespecificeerde bestanden.
De nieuwe tijd is steeds het moment van nu. Na afloop hebben alle bestanden dezelfde modification time.

De te behandelen bestanden worden gespecificeerd door:

    - de argumenten (een directory argument betekent steeds alle bestanden in de directory)
    - de :option:`recurse` vlag
    - de :option:`pattern` vlag(gen)

Indien de argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

Er moet minstens 1 bestand gespecificeerd zijn. Dit kan vb. `.` zijn (de huidige directory)

Door 1 of meerdere `--pattern` vlaggen te specificeren, kunnen toegelaten patronen op de *basename* van de bestanden worden opgelegd.
Met de `--recurse` vlag kan ook de ganse *directory tree* worden behandeld.


Opties
~~~~~~~~~~~

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--cwd=<dir>                  De te behandelen bestanden (de *sources*) worden relatief genomen tegenover `dir`

--pattern=<wildcard>         Er kunnen meeerdere dergelijke vlaggen worden gespecificeerd. 
                             Deze filteren de *sources* op hun *basename*

--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs touch *.m`

:samp:`qtechng fs touch . --pattern="*.m" --recurse`

