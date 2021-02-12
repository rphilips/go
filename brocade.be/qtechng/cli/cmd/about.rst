


:command:`qtechng about`
====================================


Synopsis
~~~~~~~~~

:dfn:`qtechng about` toont informatie over het `qtechng` bestand.
Daardoor wordt het eenvoudig om te controleren of je werkt met de laatste versie. 

Dit commando werkt op:

    - `W`: Werkstations
    - `B`: Ontwikkelmachine (dev.anet.be)
    - `P`: Productie servers


Opties
~~~~~~~~

--remote                     Voer het commando uit op de ontwikkelserver

--stdout=<filename>          Schrijf de `output` naar het vermeld bestand.
                             De naam van het output bestand is ofwel absoluut, ofwel relatief
                             tegenover de directory waarin het commando wordt uitgevoerd
                             (M.a.w. eventuele specificatie van `--cwd=...` heeft geen invloed op de plaats
                             van dit bestand)

--jq=<jsonpath>              Specificeer een `JSONPath` uitdrukking om het resultaat aan te passen


Voorbeelden
~~~~~~~~~~~~

:samp:`qtechng about`


.. code-block:: json

    {
        "host": "rphilips-XPS-17-9700",
        "time": "2021-02-10T16:46:33+01:00",
        "ERROR": null,
        "RESULT": {
            "!!uname": "rphilips-XPS-17-9700",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.02.10-16:39:12",
            "!BuildWith": "go1.16rc1"
        }
    }




