// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"net/url"

	activities_model "code.gitea.io/gitea/models/activities"
	"code.gitea.io/gitea/models/perm"
	api "code.gitea.io/gitea/modules/structs"
)

// ToNotificationThread convert a Notification to api.NotificationThread
func ToNotificationThread(n *activities_model.Notification) *api.NotificationThread {
	result := &api.NotificationThread{
		ID:        n.ID,
		Unread:    !(n.Status == activities_model.NotificationStatusRead || n.Status == activities_model.NotificationStatusPinned),
		Pinned:    n.Status == activities_model.NotificationStatusPinned,
		UpdatedAt: n.UpdatedUnix.AsTime(),
		URL:       n.APIURL(),
	}

	// since user only get notifications when he has access to use minimal access mode
	if n.Repository != nil {
		result.Repository = ToRepo(n.Repository, perm.AccessModeRead)

		// This permission is not correct and we should not be reporting it
		for repository := result.Repository; repository != nil; repository = repository.Parent {
			repository.Permissions = nil
		}
	}

	// handle Subject
	switch n.Source {
	case activities_model.NotificationSourceIssue:
		result.Subject = &api.NotificationSubject{Type: api.NotifySubjectIssue}
		if n.Issue != nil {
			result.Subject.Title = n.Issue.Title
			result.Subject.URL = n.Issue.APIURL()
			result.Subject.HTMLURL = n.Issue.HTMLURL()
			result.Subject.State = n.Issue.State()
			comment, err := n.Issue.GetLastComment()
			if err == nil && comment != nil {
				result.Subject.LatestCommentURL = comment.APIURL()
				result.Subject.LatestCommentHTMLURL = comment.HTMLURL()
			}
		}
	case activities_model.NotificationSourcePullRequest:
		result.Subject = &api.NotificationSubject{Type: api.NotifySubjectPull}
		if n.Issue != nil {
			result.Subject.Title = n.Issue.Title
			result.Subject.URL = n.Issue.APIURL()
			result.Subject.HTMLURL = n.Issue.HTMLURL()
			result.Subject.State = n.Issue.State()
			comment, err := n.Issue.GetLastComment()
			if err == nil && comment != nil {
				result.Subject.LatestCommentURL = comment.APIURL()
				result.Subject.LatestCommentHTMLURL = comment.HTMLURL()
			}

			pr, _ := n.Issue.GetPullRequest()
			if pr != nil && pr.HasMerged {
				result.Subject.State = "merged"
			}
		}
	case activities_model.NotificationSourceCommit:
		url := n.Repository.HTMLURL() + "/commit/" + url.PathEscape(n.CommitID)
		result.Subject = &api.NotificationSubject{
			Type:    api.NotifySubjectCommit,
			Title:   n.CommitID,
			URL:     url,
			HTMLURL: url,
		}
	case activities_model.NotificationSourceRepository:
		result.Subject = &api.NotificationSubject{
			Type:  api.NotifySubjectRepository,
			Title: n.Repository.FullName(),
			// FIXME: this is a relative URL, rather useless and inconsistent, but keeping for backwards compat
			URL:     n.Repository.Link(),
			HTMLURL: n.Repository.HTMLURL(),
		}
	}

	return result
}

// ToNotifications convert list of Notification to api.NotificationThread list
func ToNotifications(nl activities_model.NotificationList) []*api.NotificationThread {
	result := make([]*api.NotificationThread, 0, len(nl))
	for _, n := range nl {
		result = append(result, ToNotificationThread(n))
	}
	return result
}
