package Drone {
    doc /* Minimal example model — enough to exercise the tree, the
         structural diagram view, and live-reload. Not a conformant
         SysML v2 model; mbsecli's bundled parser only understands a
         subset of the textual notation so far. */

    part def Airframe {
        part battery : Battery;
        part esc : ESC;
        part motor : Motor;

        in port powerIn : ElectricalPort;
    }

    part def Battery {
        attribute capacity : Real;
        out port powerOut : ElectricalPort;
    }

    part def ESC {
        in port powerIn : ElectricalPort;
        out port powerOut : ElectricalPort;
    }

    part def Motor {
        in port powerIn : ElectricalPort;
    }

    action def Fly {
        action takeoff;
        action cruise;
        action land;
    }

    state def FlightMode {
        state grounded;
        state airborne;
        state landing;
    }

    requirement def MaxWeight {
        doc /* The airframe shall weigh no more than 1.2kg fully loaded. */
        attribute limitKg : Real;
    }
}
