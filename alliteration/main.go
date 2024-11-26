package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Load words dynamically for UNIX-like systems
func loadWords() ([]string, []string) {
	adjectives := []string{}
	nouns := []string{}

	// Check if the dictionary file exists
	dictionaryPath := "/usr/share/dict/words"
	file, err := os.Open(dictionaryPath)
	if err != nil {
		fmt.Printf("Error: Unable to open the dictionary file at %s.\n", dictionaryPath)
		fmt.Println("Ensure the dictionary file exists and is readable.")
		fmt.Println("For Linux, install 'wamerican' with:")
		fmt.Println("  sudo apt-get install wamerican")
		fmt.Println("For macOS, install a dictionary with Homebrew:")
		fmt.Println("  brew install wordnet")
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		// Simple filtering rules for adjectives and nouns
		if len(word) > 3 && len(word) < 12 { // Keep words of reasonable length
			if strings.HasSuffix(word, "y") || strings.HasSuffix(word, "ous") {
				adjectives = append(adjectives, word)
			} else if strings.HasSuffix(word, "er") || strings.HasSuffix(word, "ion") || strings.HasSuffix(word, "ist") {
				nouns = append(nouns, word)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the dictionary file.")
		os.Exit(1)
	}

	return adjectives, nouns
}

// Helper function: Filter words by their initial letter
func filterByInitialLetter(words []string, initial byte) []string {
	filtered := []string{}
	for _, word := range words {
		if len(word) > 0 && word[0] == initial {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// Helper function: Normalize a combination by sorting the words alphabetically
func normalizeCombination(words []string) string {
	sortedWords := append([]string{}, words...)
	sort.Strings(sortedWords)
	return strings.Join(sortedWords, " ")
}

// Helper function: Check if a word exists in a slice
func contains(slice []string, word string) bool {
	for _, w := range slice {
		if w == word {
			return true
		}
	}
	return false
}

// Generate suggestions
func generateSuggestions(name string, count int, adjectives, nouns []string, alliteration bool, maxLength int) []string {
	suggestions := []string{}
	titleCaser := cases.Title(language.English)

	if alliteration {
		initial := strings.ToLower(name)[0]
		adjectives = filterByInitialLetter(adjectives, initial)
		nouns = filterByInitialLetter(nouns, initial)

		if len(adjectives) == 0 || len(nouns) == 0 {
			fmt.Printf("No adjectives or nouns found starting with '%c'.\n", initial)
			return suggestions
		}
	}

	// Shuffle words for variety
	rand.Shuffle(len(adjectives), func(i, j int) { adjectives[i], adjectives[j] = adjectives[j], adjectives[i] })
	rand.Shuffle(len(nouns), func(i, j int) { nouns[i], nouns[j] = nouns[j], nouns[i] })

	// Store normalized combinations to ensure uniqueness
	usedCombinations := make(map[string]bool)

	for len(suggestions) < count {
		suggestion := name
		usedWords := []string{}
		length := len(name)

		// Safety counter to prevent infinite loops
		maxAttempts := 100
		attempts := 0

		// Pick a random number of words for this suggestion (1 to 5 words)
		wordLimit := rand.Intn(5) + 1 // Randomly pick between 1 and 5 words
		wordCount := 0

		for length < maxLength && attempts < maxAttempts && wordCount < wordLimit {
			attempts++

			var word string
			// 50% chance to pick an adjective
			if len(adjectives) > 0 && rand.Intn(2) == 0 {
				word = adjectives[rand.Intn(len(adjectives))]
			} else if len(nouns) > 0 {
				word = nouns[rand.Intn(len(nouns))]
			}

			// Ensure the word hasn't been used in this suggestion
			if word != "" && !contains(usedWords, word) {
				word = titleCaser.String(word)
				if length+len(word) <= maxLength { // No space to account for
					suggestion += word
					usedWords = append(usedWords, word)
					length += len(word)
					wordCount++
				}
			}

			// Break if no more words can fit
			if len(usedWords) >= len(adjectives)+len(nouns) {
				break
			}
		}

		// Normalize the combination by sorting words alphabetically
		normalized := normalizeCombination(usedWords)

		// Ensure uniqueness based on the normalized combination
		if len(suggestion) <= maxLength && !usedCombinations[normalized] {
			suggestions = append(suggestions, suggestion)
			usedCombinations[normalized] = true
		}

		// Stop if we've exhausted all possible combinations
		if len(usedCombinations) >= len(adjectives)*len(nouns) {
			break
		}
	}

	return suggestions
}

func main() {
	adjectives, nouns := loadWords()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n\nEnter a name to generate suggestions (or type 'exit' to quit):")
		nameInput, _ := reader.ReadString('\n')
		name := strings.TrimSpace(nameInput)

		if strings.ToLower(name) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if len(name) == 0 {
			fmt.Println("Please enter a valid name.")
			continue
		}

		// Ask if alliteration is wanted
		fmt.Println("Do you want alliteration? (yes/no) [default: yes]:")
		alliterationInput, _ := reader.ReadString('\n')
		alliteration := strings.TrimSpace(strings.ToLower(alliterationInput))
		alliterationEnabled := alliteration == "" || alliteration == "yes"

		// Ask for max length
		fmt.Println("Enter the maximum length [default: 32]:")
		maxLengthInput, _ := reader.ReadString('\n')
		maxLengthStr := strings.TrimSpace(maxLengthInput)
		maxLength := 32
		if maxLengthStr != "" {
			parsedLength, err := strconv.Atoi(maxLengthStr)
			if err != nil || parsedLength <= 0 {
				fmt.Println("Invalid input. Using default max length of 32.")
			} else {
				maxLength = parsedLength
			}
		}

		for {
			fmt.Print("\nGenerating suggestions...\n\n")
			suggestions := generateSuggestions(name, 20, adjectives, nouns, alliterationEnabled, maxLength)

			if len(suggestions) == 0 {
				fmt.Printf("No suggestions could be generated for '%s'.\n", name)
				break
			}

			for _, suggestion := range suggestions {
				fmt.Println(suggestion)
			}

			if len(suggestions) < 20 {
				fmt.Println("\nNo more unique suggestions can be generated for this name.")
				break
			}

			fmt.Println("\nWould you like to generate more, start a new name, or exit? (more/new/exit) [default: more]")
			choiceInput, _ := reader.ReadString('\n')
			choice := strings.TrimSpace(strings.ToLower(choiceInput))

			if choice == "" {
				choice = "more"
			}

			if choice == "new" {
				break
			} else if choice == "exit" {
				fmt.Println("Goodbye!")
				return
			} else if choice != "more" {
				fmt.Println("Invalid input. Please type 'more', 'new', or 'exit'.")
			}
		}
	}
}
