package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"git.sr.ht/~glorifiedgluer/mata/mataroa"
	"github.com/spf13/cobra"
)

func NewPostsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Manage posts",
	}

	cmd.AddCommand(NewPostsCreateCommand())
	cmd.AddCommand(NewPostsDeleteCommand())
	cmd.AddCommand(NewPostsEditCommand())
	cmd.AddCommand(NewPostsGetCommand())
	cmd.AddCommand(NewPostsListCommand())
	cmd.AddCommand(NewPostsUpdateCommand())

	return cmd
}

func NewPostsCreateCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		filePath := args[0]

		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			log.Fatalf("error creating new post: file '%s' not found\n", filePath)
		}

		post, err := mataroa.NewPost(filePath)
		if err != nil {
			log.Fatalf("error creating new post: %s\n", err)
		}

		c := mataroa.NewMataroaClient()

		resp, err := c.CreatePost(ctx, mataroa.PostsCreateResquest{
			Title:       post.Title,
			PublishedAt: post.PublishedAt,
			Body:        post.Body,
		})
		if err != nil {
			log.Fatalf("error creating new post: %s\n", err)
		}

		if resp.OK {
			fmt.Printf("created post '%s' successfully! %s\n", resp.Slug, resp.URL)
		}
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func NewPostsDeleteCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		slug := args[0]

		c := mataroa.NewMataroaClient()

		response, err := c.DeletePost(ctx, slug)
		if err != nil {
			log.Fatalf("couldn't delete post: %s", err)
		}

		if !response.OK {
			log.Fatalf("couldn't delete post: %s", response.Error)
		}

		fmt.Printf("deleted post '%s' sucessfully\n", slug)
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func NewPostsEditCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]

		c := mataroa.NewMataroaClient()

		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.Fatalf("couldn't get post '%s': %s", slug, err)
		}

		file, err := os.CreateTemp("", "mata")
		if err != nil {
			log.Fatalf("couldn't create temp file: %s", err)
		}

		_, err = file.WriteString(mataroa.PostToMarkdown(response.Post))
		if err != nil {
			log.Fatalf("couldn't write markdown to file: %s", err)
		}

		editor := os.Getenv("EDITOR")
		if len(editor) == 0 {
			log.Fatalln("couldn't edit post $EDITOR environment variable not set")
		}

		tempname := file.Name()
		defer os.Remove(tempname)

		shellCommand := exec.Command(editor, tempname)
		shellCommand.Stdin = os.Stdin
		shellCommand.Stdout = os.Stdout
		err = shellCommand.Run()
		if err != nil {
			log.Fatalf("error while spawning $EDITOR: %s", err)
		}

		post, err := mataroa.NewPost(tempname)
		if err != nil {
			log.Fatalf("couldn't read new post body from temp file: %s", err)
		}

		updateResponse, err := c.UpdatePost(ctx, slug, post)
		if err != nil {
			log.Fatal(err)
		}

		if updateResponse.OK {
			log.Printf("successfully updated post '%s'!", slug)
		} else {
			log.Fatalf("couldn't update the post '%s': %s ", slug, updateResponse.Error)
		}
	}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func NewPostsGetCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		slug := args[0]

		c := mataroa.NewMataroaClient()
		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.Fatalf("couldn't get post '%s': %s", slug, err)
		}

		if !response.OK {
			log.Fatalf("couldn't get post '%s': %s", slug, response.Error)
		}

		md := mataroa.PostToMarkdown(response.Post)
		fmt.Println(md)
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func NewPostsUpdateCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]
		filepath := args[1]

		post, err := mataroa.NewPost(filepath)
		if err != nil {
			log.Fatal(err)
		}
		c := mataroa.NewMataroaClient()

		response, err := c.UpdatePost(ctx, slug, post)
		if err != nil {
			log.Fatal(err)
		}

		if response.OK {
			log.Printf("successfully updated slug %s with file %s", slug, filepath)
		} else {
			log.Fatalf("couldn't update slug %s with file %s", slug, response.Error)
		}
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a post",
		Args:  cobra.ExactArgs(2), // TODO: Add slug flag
		Run:   run,
	}
	return cmd
}

func NewPostsListCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		c := mataroa.NewMataroaClient()

		posts, err := c.ListPosts(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for _, post := range posts {
			fmt.Printf("%s\n", post.Slug)
			fmt.Printf("%s - %s", post.Title, post.PublishedAt)
			fmt.Printf("\n\n")
		}
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		Args:  cobra.ExactArgs(0),
		Run:   run,
	}
	return cmd
}
