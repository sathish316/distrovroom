factorial :: Integer -> Integer
factorial n
  | n < 0     = error "factorial is undefined for negative integers"
  | otherwise = product [1 .. n]
