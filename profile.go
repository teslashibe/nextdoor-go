package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

const meQuery = `query Me {
  me {
    user {
      id
      name {
        givenName
        displayName
      }
      neighborhood {
        id
        name
      }
      pronouns
    }
  }
}`

// GetMe returns the authenticated user's profile.
func (c *Client) GetMe(ctx context.Context) (*Profile, error) {
	data, err := c.gql(ctx, "Me", meQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("GetMe: %w", err)
	}

	var resp meResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetMe: %w: %v", ErrRequestFailed, err)
	}

	u := resp.Me.User
	return &Profile{
		ID:               u.ID,
		DisplayName:      u.Name.DisplayName,
		GivenName:        u.Name.GivenName,
		NeighborhoodID:   u.Neighborhood.ID,
		NeighborhoodName: u.Neighborhood.Name,
		Pronouns:         u.Pronouns,
	}, nil
}

const getProfileQuery = `query GetProfile($userId: ID!) {
  get_profile(userId: $userId) {
    id
    name {
      givenName
      displayName
    }
    neighborhood {
      id
      name
    }
  }
}`

// GetProfile retrieves a user's public profile by ID.
func (c *Client) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("GetProfile: %w: userID required", ErrInvalidParams)
	}

	vars := map[string]any{"userId": userID}
	data, err := c.gql(ctx, "GetProfile", getProfileQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("GetProfile: %w", err)
	}

	var resp getProfileResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetProfile: %w: %v", ErrRequestFailed, err)
	}

	p := resp.GetProfile
	return &Profile{
		ID:               p.ID,
		DisplayName:      p.Name.DisplayName,
		GivenName:        p.Name.GivenName,
		NeighborhoodID:   p.Neighborhood.ID,
		NeighborhoodName: p.Neighborhood.Name,
	}, nil
}
