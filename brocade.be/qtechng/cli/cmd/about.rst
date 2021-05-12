


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
        "ABOUT": {
            "args": [
                "qtechng",
                "about"
            ],
            "host": "rphilips-XPS-17-9700",
            "time": "2021-05-12T10:53:47+02:00"
        },
        "DATA": {
            "!!uname": "rphilips-XPS-17-9700",
            "!!user.name": "Richard Philips",
            "!!user.username": "rphilips",
            "!BuildHost": "rphilips-XPS-17-9700",
            "!BuildTime": "2021.05.10-15:19:42",
            "!BuildWith": "go1.16.4"
        },
        "ERRORS": null
    }


:samp:`qtechng about --jsonpath='$..DATA'`

.. code:block:: json

    {
        "!!uname": "rphilips-XPS-17-9700",
        "!!user.name": "Richard Philips",
        "!!user.username": "rphilips",
        "!BuildHost": "rphilips-XPS-17-9700",
        "!BuildTime": "2021.05.10-15:19:42",
        "!BuildWith": "go1.16.4"
    }

