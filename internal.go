package nextdoor

import "encoding/json"

// gqlRequest is the JSON body sent to the GraphQL endpoint.
type gqlRequest struct {
	OperationName string          `json:"operationName,omitempty"`
	Query         string          `json:"query"`
	Variables     json.RawMessage `json:"variables,omitempty"`
}

// gqlResponse is the top-level GraphQL response envelope.
type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlErrorEntry `json:"errors,omitempty"`
}

type gqlErrorEntry struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// --- internal response shapes ---

type meResponse struct {
	Me struct {
		User struct {
			ID   string `json:"id"`
			Name struct {
				GivenName   string `json:"givenName"`
				DisplayName string `json:"displayName"`
			} `json:"name"`
			Neighborhood struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"neighborhood"`
			Pronouns string `json:"pronouns"`
		} `json:"user"`
	} `json:"me"`
}

type feedResponse struct {
	Me struct {
		PersonalizedFeed struct {
			FeedItems []feedItem `json:"feedItems"`
			NextPage  string    `json:"nextPage"`
		} `json:"personalizedFeed"`
	} `json:"me"`
}

type feedItem struct {
	Typename string   `json:"__typename"`
	Post     postNode `json:"post"`
}

type postNode struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Author  struct {
		DisplayName string `json:"displayName"`
		URL         string `json:"url"`
	} `json:"author"`
	CreatedAt struct {
		EpochSeconds float64 `json:"epochSeconds"`
	} `json:"createdAt"`
	MediaAttachments []struct {
		Typename string `json:"__typename"`
		URL      string `json:"url"`
	} `json:"mediaAttachments"`
}

type createPostResponse struct {
	CreatePostV3 struct {
		FeedPostItem struct {
			Post postNode `json:"post"`
		} `json:"feedPostItem"`
	} `json:"createPostV3"`
}

// --- reactions ---

type addReactionResponse struct {
	AddReactionToPost struct {
		Post struct {
			ReactionSummaries struct {
				Summaries []struct {
					UserReactionID string `json:"userReactionId"`
				} `json:"summaries"`
			} `json:"reactionSummaries"`
		} `json:"post"`
	} `json:"addReactionToPost"`
}

// --- comments ---

type commentNode struct {
	ID     string `json:"id"`
	Author struct {
		DisplayName string `json:"displayName"`
		URL         string `json:"url"`
	} `json:"author"`
	Body      string `json:"body"`
	CreatedAt struct {
		EpochSeconds float64 `json:"epochSeconds"`
	} `json:"createdAt"`
}

type createCommentResponse struct {
	CreateCommentV3 struct {
		Comment commentNode `json:"comment"`
	} `json:"create_comment_v3"`
}

type deleteCommentResponse struct {
	DeleteComment struct {
		Success bool `json:"success"`
	} `json:"delete_comment"`
}

type getProfileResponse struct {
	GetProfile struct {
		ID   string `json:"id"`
		Name struct {
			GivenName   string `json:"givenName"`
			DisplayName string `json:"displayName"`
		} `json:"name"`
		Neighborhood struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"neighborhood"`
	} `json:"get_profile"`
}

// --- messaging ---

type createChannelResponse struct {
	CreateRtmChannel struct {
		ChannelID string `json:"channelId"`
	} `json:"createRtmChannel"`
}

type deleteMessageResponse struct {
	DeleteRtmMessageV2 struct {
		Success bool `json:"success"`
	} `json:"deleteRtmMessageV2"`
}

// Stream Chat send-message response.
type streamSendMessageResponse struct {
	Message struct {
		ID   string `json:"id"`
		Text string `json:"text"`
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"message"`
}

// --- feed with inline comments ---

type feedWithCommentsResponse struct {
	Me struct {
		PersonalizedFeed struct {
			FeedItems []feedItemWithComments `json:"feedItems"`
		} `json:"personalizedFeed"`
	} `json:"me"`
}

type feedItemWithComments struct {
	Typename string `json:"__typename"`
	Post     struct {
		ID       string `json:"id"`
		Comments struct {
			PagedComments struct {
				Edges []struct {
					Node struct {
						Comment commentNode `json:"comment"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"pagedComments"`
			TotalCommentCount int `json:"totalCommentCount"`
		} `json:"comments"`
	} `json:"post"`
}

// --- search (searchResultView is a union array) ---

type searchPostFeedResponse struct {
	SearchPostFeed struct {
		SearchResultView []json.RawMessage `json:"searchResultView"`
	} `json:"searchPostFeed"`
}

type searchNeighborFeedResponse struct {
	SearchNeighborFeed struct {
		SearchResultView []json.RawMessage `json:"searchResultView"`
	} `json:"searchNeighborFeed"`
}

type searchResultSection struct {
	Typename          string `json:"__typename"`
	SearchResultItems struct {
		Edges    []searchEdge `json:"edges"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"searchResultItems"`
}

type searchEdge struct {
	Node searchResultNode `json:"node"`
}

type searchResultNode struct {
	Title struct {
		Text string `json:"text"`
	} `json:"title"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	URL       string `json:"url"`
	ContentID string `json:"contentId"`
}

// --- notifications ---

type notificationFeedResponse struct {
	Me struct {
		NotificationFeed struct {
			BadgeCount int                `json:"badgeCount"`
			FeedItems  []notificationItem `json:"feedItems"`
		} `json:"notificationFeed"`
	} `json:"me"`
}

type notificationItem struct {
	Typename     string           `json:"__typename"`
	Notification notificationNode `json:"notification"`
}

type notificationNode struct {
	ID   string `json:"id"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	IsRead    bool `json:"isRead"`
	CreatedAt struct {
		EpochSeconds float64 `json:"epochSeconds"`
	} `json:"createdAt"`
}

// --- Stream Chat (DM history) ---

type streamChannelQueryResponse struct {
	Channel struct {
		ID string `json:"id"`
	} `json:"channel"`
	Messages []streamMessage `json:"messages"`
}

type streamMessage struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}

type streamChannelListResponse struct {
	Channels []struct {
		Channel struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"channel"`
		Messages []streamMessage `json:"messages"`
		Members  []struct {
			User struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"members"`
	} `json:"channels"`
}
