package rep

/*
    Imports
*/

import (
  "fmt"
  "os"
  "strconv"
  "text/tabwriter"
)

/*
    Global Variables
*/

var writer *tabwriter.Writer

/*
    Initialization Function
*/

func init() {

  writer = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)

}

/*
    Functions
*/

func (table *ReputationTable) Log() {

  fmt.Println("********* Reputations *********")
  fmt.Fprintln(writer, "Peer \t Sig-Rep \t Contrib-Rep")

  table.mutex.Lock()

  for peer, rep := range table.sigReps {

    contribRep, ok := table.contribReps[peer]

    var contribRepLog string
    if ok {
      contribRepLog = strconv.FormatFloat(float64(contribRep), 'f', -1, 32)
    } else {
      contribRepLog = "N/A"
    }

    fmt.Fprintln(writer, peer.Identifier, "\t", rep, "\t", contribRepLog)

  }

  for peer, rep := range table.contribReps {

    if _, ok := table.sigReps[peer] ; !ok {

      fmt.Fprintln(writer, peer.Identifier, "\tN/A\t", rep)

    }

  }

  table.mutex.Unlock()

  writer.Flush()

  fmt.Println("*******************************")

}
