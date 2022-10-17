package vcd

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	rdpStableMaxAttempts                = 200
	rdpStableSuccessfulAttemptsRequired = 40
	rdpStableInterval                   = 5 * time.Second
)

// The Windows VMs take a bit of time to become available (as the VMware Tools reboot them serveral times)
func (a *VcdDriver) waitForRDPStable(ip string) error {
	log.Debug().Msgf("Waiting for RDP to be stable")

	attempts := 0
	successfulAttempts := 0

	for {
		attempts++
		if attempts >= rdpStableMaxAttempts {
			return fmt.Errorf("failed to connect via RDP")
		}

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:3389", ip), 3*time.Second)
		if err != nil {
			if successfulAttempts > 0 {
				log.Debug().Msgf("RDP was available, but it is not anymore...")
				// successfulAttempts = 0
			} else {
				log.Debug().Msgf("Nothing listening in %s:3389 yet (attempt %d out of %d)", ip, attempts, rdpStableMaxAttempts)
			}

			time.Sleep(rdpStableInterval)
			continue
		}

		successfulAttempts++

		log.Debug().Msgf("Connected to %s:3389. Successful attept %d out of %d required. Total attempts %d, max %d",
			ip, successfulAttempts, rdpStableSuccessfulAttemptsRequired,
			attempts, rdpStableMaxAttempts)

		if successfulAttempts >= rdpStableSuccessfulAttemptsRequired {
			return nil
		}

		conn.Close()
		time.Sleep(rdpStableInterval)
	}
}
