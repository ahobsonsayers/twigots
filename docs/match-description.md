I would like to implement fuzzy substring matching (similarity scoring) for the purpose of checking if a substring is within another string, allowing for things like minor misspelling

Here are the constraints:

- I cannot rely solely on substring matching because if there is a misspelling in the wanted event name (e.g., "Taylar Swift" or "Taylor Swft"), I still want it to match. I will set a minimum confidence threshold for a match to occur.
- I cannot use Levenshtein distance because the distance between the two event names in the example above is too large.
- I cannot use the Smith-Waterman algorithm because I want to ensure that only the full substring matches. For instance, "Taylor Swift: The Era Tour" should not match "Taylor Swift" if "Taylor Swift" is not a substring of the actual event name.

As a note: Before any matching or calculation is done, the string will be normalized by converting it to lowercase and removing any symbols. This preprocessing step is already accounted for and does not need to be considered in the matching process.

Example 1:

- Substring: Taylor Swift
- Target: Taylor Swift: The Eras Tour
- Should give a similarity score of 1

Example 2:

- Substring: Tayler Swift
- Target: Taylor Swift: The Eras Tour
- Should give a high similarity score, as teh substring is only one letter off

Example 3:

- Substring: Taylor Swift: The Eras Tour
- Target: Taylor Swift
- Should give a low similarity score, as the substring is not in the target

My current solution is to use a modified version of Smith-Waterman, with Levenshtein distance done on a word level. The full substring must appear, with penalties for insertions in the middle of the substring found in the target.

This is my implementation. How can i improve it?

Currently:

- Substring: Taylor Swift
- Target: Taylor The Swift
- Gives a perfect score of 1, when it should be below one
