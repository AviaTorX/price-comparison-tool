**Tech** **task**

Create a generic tool that can fetch price of a given product from
multiple websites based on the country the consumer is shopping from. It
should ideally fetch rates from all websites selling the product for a
given country, and ensure that the product matches with the requirement.

Example input:

> • {"country": "US", "query":"iPhone 16 Pro, 128GB"}
>
> • {"country": "IN", "query": "boAt Airdopes 311 Pro"}

Note that, we expect the tool to work **across** **ALL** **countries**
for **EVERY** **category** of products typically sold online. We expect
the tool to be, therefore, capable of looking for prices from **ANY**
**applicable** **website**.

We expect the tool to produce list of results ranked in ascending of
price: Sample output:

> \[
>
> {
>
> "link":"https://apple.in/...", "price":"999", "currency":"USD",
>
> "productName":"Apple iPhone 16 Pro", "parameter1" : .....
>
> }, ...
>
> \]
>
> **Evaluation** **criteria:**

Your solution will be judged on all of the following \[no-speciﬁc order
treat all points as equally signiﬁcant\]:

> •
>
> •
>
> •

**<u>Accuracy</u>** **&** **Reliability:** Any result produced should be
accurate. The parsed data like productName and price must match with
what is given in the source url. Further only the products that actually
match the user's query should be fetched.

**Coverage:** The tool will be judged on the fact that can it fetch
information for all types of products, not just limited to few category
of products or few popular websites or 2-3 countries.

**Quality:** The solution will be judged on the quality of results
produced. For example, mobile phones may be cheaper on websites like
sangeetha mobiles that might be more relevant in the country provided
\[not always true, depends on the product\].

P.S. while it's not compulsory, but brownie points for using LLMs/AI to
create a better tool.

> **Submission** **instructions:**
>
> •

Hosted URL. Please host your solution in any online platform (like
Vercel) and provide a URL for us to test. Frontend is not mandatory but
preferred if possible. If frontend is not developed, then ensure the
github repo includes **working** **curl** request(s) for us to test. The
curl should be mentioned in the readme section of the repo properly.

• Github repo link (public repo)

> The repo **must** include complete instructions to test \[including
> all dependencies\]. Preferred
>
> dockerised applications to prevent issues arising due to dependency
> mismatch. While we will attempt our best to ensure no errors occur due
> to our testing environment (even if your testing instructions are
> insuﬃcient), we will reject applications which fail tests.
>
> •
>
> •

The repo must have proof of working for query: {"country": "US",
"query":"iPhone 16 Pro, 128GB"} -video or image.

Working example curl request if no frontend for the tool.

• Use this form to submit your application:
[<u>https://forms.gle/QLFB7zD15JncCHnJ7</u>](https://forms.gle/QLFB7zD15JncCHnJ7)
