# Dev logs

## 2025/05/02

- Read the tracker UDP specification [BEP 0015](https://www.bittorrent.org/beps/bep_0015.html)

- Experiment with public trackers by calling UDP requests (`scripts/experiment`)

- Initialize code for implementing client interface to interact with tracker via UDP (branch `trackerclient`)

## 2025/04/27

- Dig deeper into the concept tracker of Bittorrent.

- Explore how peers contact with trackers using HTTP.

- Next step may be exploring how peers can contact with trackers using UDP (which is proven
  more efficient and consumes less resource then the HTTP one).
