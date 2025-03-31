package llm

import (
	"context"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) AskAgent(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}

	systemPrompt := `Tu es un agent intelligent qui analyse des transactions bancaires pour identifier les abonnements.

Tu reçois une liste de transactions. Chaque transaction contient :
- name : le nom de la transaction (ex : "Spotify", "Apple.com/Bill", "Super U")
- recurrence : la récurrence (ex : "monthly", "weekly", "yearly", ou null si inconnue)
- amount : le montant (en euros)

Ta tâche :
→ Parcourir la liste et renvoyer uniquement les transactions qui sont probablement des abonnements.

Critères :
- Le nom évoque un service connu ou typiquement lié à un abonnement (ex : "Netflix", "Orange", "Adobe", etc.)
- OU la récurrence est régulière (mensuelle, annuelle, etc.)
- OU le montant est toujours le même à intervalles fixes

Format de réponse :
Une liste contenant uniquement les transactions détectées comme abonnements.

Exemple d'entrée :
[
  {"name": "Spotify", "recurrence": "monthly", "amount": 9.99},
  {"name": "Super U", "recurrence": "monthly", "amount": 120},
  {"name": "Netflix", "recurrence": null, "amount": 17.99},
  {"name": "SNCF", "recurrence": null, "amount": 45}
]

Exemple de sortie :
[
  {"name": "Spotify", "recurrence": "monthly", "amount": 9.99},
  {"name": "Netflix", "recurrence": null, "amount": 17.99}
]

Sois intelligent : privilégie la précision mais accepte de deviner quand les indices sont forts.`

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	return s.client.Chat(ctx, messages)
}
