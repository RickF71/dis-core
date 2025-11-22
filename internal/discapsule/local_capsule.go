package discapsule

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

// LocalCapsule is a minimal, interactive implementation of Capsule that
// prompts on stdin/stdout. It's the simplest possible human-in-the-loop
// authority source for BedrockBootstrap.
type LocalCapsule struct{}

func NewLocalCapsule() *LocalCapsule {
	return &LocalCapsule{}
}

func (c *LocalCapsule) PerformBedrockAuth(ch BedrockChallenge) (BedrockGrant, error) {
	fmt.Println("🪨 DIS-Core BedrockBootstrap")
	fmt.Println("  DIS-Core requires human approval to establish the 1D root domain `null`.")
	fmt.Printf("  Instance: %s\n", ch.InstanceID)
	fmt.Printf("  Nonce:    %s\n", ch.Nonce)
	fmt.Printf("  Time:     %s\n", ch.Timestamp)
	fmt.Print("  Approve bedrock creation? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if line != "y" && line != "Y" {
		return BedrockGrant{}, fmt.Errorf("bedrock declined by human")
	}

	return BedrockGrant{
		GrantID:    uuid.New().String(),
		InstanceID: ch.InstanceID,
		Nonce:      ch.Nonce,
		Approved:   true,
	}, nil
}
