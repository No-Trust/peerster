package rep

/*
    Imports
*/

import (
  "fmt"
  "os"
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

  fmt.Println("\n************ Reputations ************\n")

  table.mutex.Lock()

  fmt.Println("Signature-based Reputations:")
  fmt.Fprintln(writer, "Peer \t Sig-Rep")

  for peer, rep := range table.sigReps {

    fmt.Fprintln(writer, peer, "\t", rep)

  }

  writer.Flush()

  fmt.Println("\nContribution-based Reputations:")
  fmt.Fprintln(writer, "Peer \t Contrib-Rep")

  for peer, rep := range table.contribReps {

    fmt.Fprintln(writer, peer, "\t", rep)

  }

  writer.Flush()

  table.mutex.Unlock()

  fmt.Println("\n*************************************\n")

}
