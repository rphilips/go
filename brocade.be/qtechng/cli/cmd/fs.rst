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


:command:`qtechng fs copy`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs copy` kopieert lokale bestanden.
Het is ook mogelijk om de originele bestanden te verwijderen waardoor het commando
effectief ook een `rename` functie heeft.

Het eerste argument is een *zoek-string*.
Het tweede argument is een *vervang-string*.
De *zoek-string* wordt opgezocht in de *absolute pathname* van de behandelde bestanden (de *sources*). 
De zoekstring wordt vervangen door de *vervang-string* en geeft de naam van het target bestand.

De *zoek-string* kan ook een reguliere uitdrukking zijn (specificeer de `--regexp` vlag) 
en de *vervang-string* kan dan deelresultaten opnemen.

Vervolgens wordt de *source* gekopieerd naar de *target*

Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

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

--regexp                     De zoek-string is een reguliere uitdrukking. 
                             De vervang-string kan gebruik maken `$0`, `$1`, ... constructies

--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--delete                     Eens gekopieerd, worden de *sources* ook geschrapt. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs copy d macro isad.d thing.d`

:samp:`qtechng fs copy pdf PDF . --recurse`

:samp:`qtechng fs copy '.*' '${0}.bak' . --regexp --recurse --pattern='*.txt'`




:command:`qtechng fs grep`
---------------------------------

Synopsis
~~~~~~~~~

:dfn:`qtechng fs grep` zoekt een string in lokale bestanden.

Het eerste argument is een *zoek-string*.

De *zoek-string* wordt lijn per lijn gezocht in de gespecificeerde bestanden

De *zoek-string* kan ook een reguliere uitdrukking zijn (specificeer de `--regexp` vlag) 

De `--tolower` vlag kan er voor zorgen dat de lijnen worden omgevormd naar *lowercase* karakters.


Indien argumenten ontbreken gaat de software een dialoog met de gebruiker opstarten.


Hoe worden de *sources* geselecteerd ?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

De te behandelen bestanden zijn de andere argumenten.
Dit kunnen ook directories zijn: alle reguliere bestanden worden behandeld.

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

--regexp                     De zoek-string is een reguliere uitdrukking. 
                             De vervang-string kan gebruik maken `$0`, `$1`, ... constructies

--recurse                    Alle reguliere bestanden, stroomafwaarts van de de argumenten die directories zijn
                             worden behandeld.

--tolower                    De hoofdletters worden omgezet naar kleine letters. 

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~~

:samp:`qtechng fs grep m4_CO . --recurse --pattern='*.m'`




