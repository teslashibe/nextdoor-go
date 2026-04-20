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
	} `json:"mediaAttachments"`
}

type createPostResponse struct {
	CreatePostV3 struct {
		Post postNode `json:"post"`
	} `json:"createPostV3"`
}

type deletePostResponse struct {
	DeletePost struct {
		Success bool `json:"success"`
	} `json:"deletePost"`
}

type addReactionResponse struct {
	AddReactionToPost struct {
		Success bool `json:"success"`
	} `json:"addReactionToPost"`
}

type removeReactionResponse struct {
	RemoveReactionFromPost struct {
		Success bool `json:"success"`
	} `json:"removeReactionFromPost"`
}

type pagedCommentsResponse struct {
	PagedComments struct {
		Comments []commentNode `json:"comments"`
		NextPage string        `json:"nextPage"`
	} `json:"pagedComments"`
}

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

type chatListResponse struct {
	Chats []struct {
		ID           string   `json:"id"`
		Participants []string `json:"participants"`
	} `json:"chats"`
}

type searchPostFeedResponse struct {
	SearchPostFeed struct {
		Results []searchNode `json:"results"`
	} `json:"searchPostFeed"`
}

type searchNeighborsResponse struct {
	SearchNeighbor struct {
		Results []searchNode `json:"results"`
	} `json:"searchNeighbor"`
}

type searchNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type notificationFeedResponse struct {
	NotificationFeed struct {
		Notifications []notificationNode `json:"notifications"`
	} `json:"notificationFeed"`
}

type notificationNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Link  string `json:"link"`
	Read  bool   `json:"read"`
	CreatedAt struct {
		EpochSeconds float64 `json:"epochSeconds"`
	} `json:"createdAt"`
}
