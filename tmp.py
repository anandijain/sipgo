# for quick copying
import pandas as pd

df = pd.read_csv('lines2.csv', index_col=False)
df2 = pd.read_csv('scores2.csv', index_col=False)
b = pd.merge(df, df2, on="game_id")
